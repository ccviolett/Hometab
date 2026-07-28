//go:build darwin

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchAgentDoesNotOpenBrowser(t *testing.T) {
	plist := renderLaunchAgentPlist(macServicePaths{
		BinPath:     "/Applications/Hometab/hometab",
		StdoutPath:  "/tmp/hometab.log",
		StderrPath:  "/tmp/hometab.err.log",
		Target:      "gui/501",
		ServiceName: "gui/501/" + launchAgentLabel,
	})

	if !strings.Contains(plist, "<string>--no-open</string>") {
		t.Fatal("LaunchAgent must disable automatic browser opening")
	}
}

func TestServicePathsUseHometabForNewInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	paths, err := servicePaths()
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(paths.AppDir) != "Hometab" {
		t.Fatalf("AppDir = %q, want Hometab directory", paths.AppDir)
	}
	if filepath.Base(paths.LogDir) != "Hometab" {
		t.Fatalf("LogDir = %q, want Hometab directory", paths.LogDir)
	}
	if filepath.Base(paths.PlistPath) != launchAgentLabel+".plist" {
		t.Fatalf("PlistPath = %q, want current label", paths.PlistPath)
	}
}

func TestServicePathsFindExistingLegacyInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	plist := filepath.Join(home, "Library", "LaunchAgents", legacyLaunchAgentLabel+".plist")
	if err := os.MkdirAll(filepath.Dir(plist), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plist, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	paths, err := servicePaths()
	if err != nil {
		t.Fatal(err)
	}

	if paths.PlistPath != plist {
		t.Fatalf("PlistPath = %q, want %q", paths.PlistPath, plist)
	}
	if filepath.Base(paths.AppDir) != "Home" {
		t.Fatalf("AppDir = %q, want legacy Home directory", paths.AppDir)
	}
}
