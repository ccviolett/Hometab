//go:build darwin

package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureRegistersWithoutStartingAndDisablesWithoutStopping(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source-hometab")
	if err := os.WriteFile(source, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	paths := testServicePaths(root)
	launchctlCalls := 0
	manager := &darwinManager{
		paths:             func() (servicePaths, error) { return paths, nil },
		currentExecutable: func() (string, error) { return source, nil },
		loaded:            func(string) bool { return false },
		launchctl: func(...string) error {
			launchctlCalls++
			return nil
		},
	}

	status, err := manager.Configure(Config{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !status.Enabled || status.Active {
		t.Fatalf("unexpected enabled status: %+v", status)
	}
	if launchctlCalls != 0 {
		t.Fatalf("Configure called launchctl %d times", launchctlCalls)
	}
	content, err := os.ReadFile(paths.PlistPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "<string>--no-open</string>") {
		t.Fatal("registered LaunchAgent must disable browser opening by default")
	}
	copied, err := os.ReadFile(paths.BinPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(copied) != "binary" {
		t.Fatalf("copied binary = %q", copied)
	}

	status, err = manager.Configure(Config{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if status.Enabled {
		t.Fatalf("unexpected disabled status: %+v", status)
	}
	if _, err := os.Stat(paths.BinPath); err != nil {
		t.Fatalf("disable removed the installed binary: %v", err)
	}
}

func TestResolveServicePathsUsesCurrentAndLegacyNames(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	paths, err := resolveServicePaths()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(paths.AppDir) != "Hometab" {
		t.Fatalf("AppDir = %q", paths.AppDir)
	}
	if filepath.Base(paths.PlistPath) != launchAgentLabel+".plist" {
		t.Fatalf("PlistPath = %q", paths.PlistPath)
	}

	legacyPlist := filepath.Join(home, "Library", "LaunchAgents", legacyLaunchAgentLabel+".plist")
	if err := os.MkdirAll(filepath.Dir(legacyPlist), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPlist, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	paths, err = resolveServicePaths()
	if err != nil {
		t.Fatal(err)
	}
	if paths.PlistPath != legacyPlist || filepath.Base(paths.AppDir) != "Home" {
		t.Fatalf("legacy paths not preserved: %+v", paths)
	}
}

func testServicePaths(root string) servicePaths {
	appDir := filepath.Join(root, "Hometab")
	return servicePaths{
		AppDir:      appDir,
		BinDir:      filepath.Join(appDir, "bin"),
		BinPath:     filepath.Join(appDir, "bin", "hometab"),
		LogDir:      filepath.Join(root, "logs"),
		StdoutPath:  filepath.Join(root, "logs", "hometab.log"),
		StderrPath:  filepath.Join(root, "logs", "hometab.err.log"),
		PlistPath:   filepath.Join(root, "LaunchAgents", launchAgentLabel+".plist"),
		Target:      "gui/501",
		ServiceName: "gui/501/" + launchAgentLabel,
	}
}
