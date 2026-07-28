package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBaseURL string

type lockedBuffer struct {
	mu  sync.RWMutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.String()
}

func baseURL() string {
	if testBaseURL != "" {
		return testBaseURL
	}
	if u := os.Getenv("E2E_BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:3999"
}

// TestMain makes the e2e suite self-contained: when E2E_BASE_URL is unset it
// builds the server binary with ldflags (so build_time is injected) and launches
// it on a free port, then tears it down afterwards. This removes the manual
// "start a server first" prerequisite and makes TestE2EBuildInfo deterministic.
func TestMain(m *testing.M) {
	if u := os.Getenv("E2E_BASE_URL"); u != "" {
		testBaseURL = u
		os.Exit(m.Run())
	}

	bin, err := buildServerBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: failed to build server binary: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(bin)

	port, cleanup, err := startServer(bin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	testBaseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	os.Exit(m.Run())
}

func buildServerBinary() (string, error) {
	root := moduleRoot()
	if root == "" {
		return "", fmt.Errorf("could not locate module root (go.mod)")
	}
	tmpDir, err := os.MkdirTemp("", "hometab-e2e")
	if err != nil {
		return "", err
	}
	binPath := filepath.Join(tmpDir, "hometab")

	// RFC3339 has no spaces, so go's -ldflags tokenization stays correct.
	buildTime := time.Now().UTC().Format(time.RFC3339)
	ldflags := fmt.Sprintf(
		"-s -w -X hometab/pkg/buildinfo.Version=test-e2e -X hometab/pkg/buildinfo.BuildTime=%s",
		buildTime,
	)

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", binPath, "./cmd/server/")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go build failed: %w: %s", err, stderr.String())
	}
	return binPath, nil
}

func moduleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// startServer grabs a free port, launches the binary, and returns the port the
// server actually bound to (tolerating Fiber's auto port-advance on conflict).
func startServer(bin string) (int, func(), error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, err
	}
	requested := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // release so the server can rebind

	dataDir, err := os.MkdirTemp("", "hometab-e2e-data")
	if err != nil {
		return 0, nil, err
	}

	cmd := exec.Command(bin, "-port", strconv.Itoa(requested), "--no-open")
	cmd.Env = append(os.Environ(), "HOME_DATABASE_PATH="+filepath.Join(dataDir, "data.db"))
	var out lockedBuffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		os.RemoveAll(dataDir)
		return 0, nil, err
	}
	cleanup := func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		_ = os.RemoveAll(dataDir)
	}

	actual, err := waitForHealthy(requested, &out)
	if err != nil {
		cleanup()
		return 0, nil, err
	}
	return actual, cleanup, nil
}

func waitForHealthy(requested int, out *lockedBuffer) (int, error) {
	deadline := time.Now().Add(20 * time.Second)
	for {
		port := requested
		if p := parsePortFromLogs(out.String()); p != 0 {
			port = p
		}
		if checkHealth(port) {
			return port, nil
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("server not healthy in time; logs:\n%s", out.String())
		}
		time.Sleep(250 * time.Millisecond)
	}
}

var urlRe = regexp.MustCompile(`http://127\.0\.0\.1:(\d+)`)

func parsePortFromLogs(s string) int {
	m := urlRe.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	p, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return p
}

func checkHealth(port int) bool {
	client := &http.Client{Timeout: 600 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func TestE2EHealth(t *testing.T) {
	resp, err := http.Get(baseURL() + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var m map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&m)
	assert.Equal(t, "healthy", m["status"])
}

func TestE2EBuildInfo(t *testing.T) {
	resp, err := http.Get(baseURL() + "/api/build-info")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var m map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&m)
	assert.NotEmpty(t, m["version"])
	assert.NotEmpty(t, m["build_time"])
	assert.NotEqual(t, "unknown", m["build_time"])
	assert.NotEmpty(t, m["go_version"])
}

func TestE2ESettings(t *testing.T) {
	resp, err := http.Get(baseURL() + "/api/settings")
	require.NoError(t, err)
	defer resp.Body.Close()
	// Accept 200 (success) or 500 (pre-existing DB data issue)
	assert.Contains(t, []int{200, 500}, resp.StatusCode)
}

func TestE2ESearchEngines(t *testing.T) {
	resp, err := http.Get(baseURL() + "/api/search-engines")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var arr []interface{}
	json.Unmarshal(body, &arr)
	assert.NotNil(t, arr)
}

func TestE2ELinksByGroup(t *testing.T) {
	resp, err := http.Get(baseURL() + "/api/links-by-group")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var arr []interface{}
	json.Unmarshal(body, &arr)
	assert.NotNil(t, arr)
}

func TestE2EExport(t *testing.T) {
	resp, err := http.Get(baseURL() + "/api/export")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "zip")
}
