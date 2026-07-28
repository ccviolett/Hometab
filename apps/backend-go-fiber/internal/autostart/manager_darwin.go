//go:build darwin

package autostart

import (
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	launchAgentLabel       = "com.species.hometab"
	legacyLaunchAgentLabel = "com.powerbase.home"
)

type darwinManager struct {
	paths             func() (servicePaths, error)
	currentExecutable func() (string, error)
	loaded            func(string) bool
	launchctl         func(...string) error
}

type servicePaths struct {
	AppDir      string
	BinDir      string
	BinPath     string
	LogDir      string
	StdoutPath  string
	StderrPath  string
	PlistPath   string
	Target      string
	ServiceName string
}

func newManager() Manager {
	return &darwinManager{
		paths:             resolveServicePaths,
		currentExecutable: resolveCurrentExecutable,
		loaded:            launchAgentLoaded,
		launchctl:         runLaunchctl,
	}
}

func (m *darwinManager) Status() (Status, error) {
	paths, err := m.paths()
	if err != nil {
		return Status{}, err
	}
	status := Status{
		Platform:  runtime.GOOS,
		Supported: true,
		Active:    m.loaded(paths.ServiceName),
	}
	_, err = os.ReadFile(paths.PlistPath)
	if errors.Is(err, os.ErrNotExist) {
		return status, nil
	}
	if err != nil {
		return Status{}, err
	}
	status.Enabled = true
	return status, nil
}

func (m *darwinManager) Configure(config Config) (Status, error) {
	paths, err := m.paths()
	if err != nil {
		return Status{}, err
	}
	if config.Enabled {
		err = m.register(paths)
	} else {
		err = removeIfExists(paths.PlistPath)
	}
	if err != nil {
		return Status{}, err
	}
	return m.Status()
}

func (m *darwinManager) Install() error {
	paths, err := m.paths()
	if err != nil {
		return err
	}
	if err := m.register(paths); err != nil {
		return err
	}
	return m.start(paths)
}

func (m *darwinManager) Uninstall() error {
	paths, err := m.paths()
	if err != nil {
		return err
	}
	if err := m.stop(paths); err != nil {
		return err
	}
	if err := removeIfExists(paths.PlistPath); err != nil {
		return err
	}
	return removeIfExists(paths.BinPath)
}

func (m *darwinManager) Start() error {
	paths, err := m.paths()
	if err != nil {
		return err
	}
	return m.start(paths)
}

func (m *darwinManager) Stop() error {
	paths, err := m.paths()
	if err != nil {
		return err
	}
	return m.stop(paths)
}

func (m *darwinManager) register(paths servicePaths) error {
	for _, dir := range []string{paths.BinDir, paths.LogDir, filepath.Dir(paths.PlistPath)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	src, err := m.currentExecutable()
	if err != nil {
		return err
	}
	if err := copyExecutable(src, paths.BinPath); err != nil {
		return err
	}
	return writeFileAtomic(paths.PlistPath, []byte(renderLaunchAgentPlist(paths)), 0o644)
}

func (m *darwinManager) start(paths servicePaths) error {
	if _, err := os.Stat(paths.PlistPath); err != nil {
		return err
	}
	if m.loaded(paths.ServiceName) {
		return m.launchctl("kickstart", "-k", paths.ServiceName)
	}
	return m.launchctl("bootstrap", paths.Target, paths.PlistPath)
}

func (m *darwinManager) stop(paths servicePaths) error {
	if m.loaded(paths.ServiceName) {
		return m.launchctl("bootout", paths.ServiceName)
	}
	if _, err := os.Stat(paths.PlistPath); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	err := m.launchctl("bootout", paths.Target, paths.PlistPath)
	if err != nil && !isLaunchAgentNotLoaded(err) {
		return err
	}
	return nil
}

func resolveServicePaths() (servicePaths, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return servicePaths{}, err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return servicePaths{}, err
	}

	label := launchAgentLabel
	appName := "Hometab"
	legacyPlist := filepath.Join(homeDir, "Library", "LaunchAgents", legacyLaunchAgentLabel+".plist")
	if _, err := os.Stat(legacyPlist); err == nil {
		label = legacyLaunchAgentLabel
		appName = "Home"
	}
	appDir := filepath.Join(configDir, appName)
	logDir := filepath.Join(homeDir, "Library", "Logs", appName)
	target := fmt.Sprintf("gui/%d", os.Getuid())

	return servicePaths{
		AppDir:      appDir,
		BinDir:      filepath.Join(appDir, "bin"),
		BinPath:     filepath.Join(appDir, "bin", "hometab"),
		LogDir:      logDir,
		StdoutPath:  filepath.Join(logDir, "hometab.log"),
		StderrPath:  filepath.Join(logDir, "hometab.err.log"),
		PlistPath:   filepath.Join(homeDir, "Library", "LaunchAgents", label+".plist"),
		Target:      target,
		ServiceName: target + "/" + label,
	}, nil
}

func resolveCurrentExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func copyExecutable(src, dst string) error {
	src, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	if src == dst {
		return os.Chmod(dst, 0o755)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

func writeFileAtomic(path string, content []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func renderLaunchAgentPlist(paths servicePaths) string {
	arguments := fmt.Sprintf("    <string>%s</string>\n    <string>--no-open</string>", html.EscapeString(paths.BinPath))
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
%s
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>%s</string>
  <key>StandardErrorPath</key>
  <string>%s</string>
</dict>
</plist>
`, html.EscapeString(strings.TrimPrefix(paths.ServiceName, paths.Target+"/")), arguments, html.EscapeString(paths.StdoutPath), html.EscapeString(paths.StderrPath))
}

func isLaunchAgentNotLoaded(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "No such process") ||
		strings.Contains(message, "Could not find specified service") ||
		strings.Contains(message, "service is not loaded")
}

func launchAgentLoaded(serviceName string) bool {
	return exec.Command("launchctl", "print", serviceName).Run() == nil
}

func runLaunchctl(args ...string) error {
	cmd := exec.Command("launchctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl %v: %w: %s", args, err, string(output))
	}
	return nil
}
