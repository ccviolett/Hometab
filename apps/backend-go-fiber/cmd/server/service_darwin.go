//go:build darwin

package main

import (
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	launchAgentLabel       = "com.species.hometab"
	legacyLaunchAgentLabel = "com.powerbase.home"
)

func runServiceCommand(action string) error {
	paths, err := servicePaths()
	if err != nil {
		return err
	}

	switch action {
	case "install":
		return installLaunchAgent(paths)
	case "uninstall":
		return uninstallLaunchAgent(paths)
	case "start":
		return startLaunchAgent(paths)
	case "stop":
		return stopLaunchAgent(paths)
	case "status":
		return printLaunchAgentStatus(paths)
	default:
		return fmt.Errorf("unsupported service command: %s", action)
	}
}

type macServicePaths struct {
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

func servicePaths() (macServicePaths, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return macServicePaths{}, err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return macServicePaths{}, err
	}

	label := launchAgentLabel
	appName := "Hometab"
	legacyPlist := filepath.Join(homeDir, "Library", "LaunchAgents", legacyLaunchAgentLabel+".plist")
	if _, err := os.Stat(legacyPlist); err == nil {
		label = legacyLaunchAgentLabel
		appName = "Home"
	}
	appDir := filepath.Join(configDir, appName)
	binDir := filepath.Join(appDir, "bin")
	logDir := filepath.Join(homeDir, "Library", "Logs", appName)
	target := fmt.Sprintf("gui/%d", os.Getuid())

	return macServicePaths{
		AppDir:      appDir,
		BinDir:      binDir,
		BinPath:     filepath.Join(binDir, "hometab"),
		LogDir:      logDir,
		StdoutPath:  filepath.Join(logDir, "hometab.log"),
		StderrPath:  filepath.Join(logDir, "hometab.err.log"),
		PlistPath:   filepath.Join(homeDir, "Library", "LaunchAgents", label+".plist"),
		Target:      target,
		ServiceName: target + "/" + label,
	}, nil
}

func installLaunchAgent(paths macServicePaths) error {
	if err := os.MkdirAll(paths.BinDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.LogDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.PlistPath), 0755); err != nil {
		return err
	}

	if err := stopLaunchAgent(paths); err != nil {
		return err
	}
	if err := copyCurrentExecutable(paths.BinPath); err != nil {
		return err
	}
	if err := os.WriteFile(paths.PlistPath, []byte(renderLaunchAgentPlist(paths)), 0644); err != nil {
		return err
	}
	if err := startLaunchAgent(paths); err != nil {
		return err
	}

	fmt.Printf("Hometab installed and started.\n")
	fmt.Printf("Binary: %s\n", paths.BinPath)
	fmt.Printf("LaunchAgent: %s\n", paths.PlistPath)
	fmt.Printf("Logs: %s\n", paths.LogDir)
	fmt.Printf("URL: http://127.0.0.1:52173\n")
	return nil
}

func uninstallLaunchAgent(paths macServicePaths) error {
	if err := stopLaunchAgent(paths); err != nil {
		return err
	}
	if err := os.Remove(paths.PlistPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove(paths.BinPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	fmt.Printf("Hometab service uninstalled.\n")
	fmt.Printf("User data is preserved in: %s\n", paths.AppDir)
	return nil
}

func startLaunchAgent(paths macServicePaths) error {
	if _, err := os.Stat(paths.PlistPath); err != nil {
		return err
	}
	if launchAgentLoaded(paths.ServiceName) {
		if err := runLaunchctl("kickstart", "-k", paths.ServiceName); err != nil {
			return err
		}
		fmt.Printf("Hometab service restarted.\n")
		return nil
	}
	if err := runLaunchctl("bootstrap", paths.Target, paths.PlistPath); err != nil {
		return err
	}
	fmt.Printf("Hometab service started.\n")
	return nil
}

func stopLaunchAgent(paths macServicePaths) error {
	if launchAgentLoaded(paths.ServiceName) {
		if err := runLaunchctl("bootout", paths.ServiceName); err != nil {
			return err
		}
		fmt.Printf("Hometab service stopped.\n")
		return nil
	}
	if _, err := os.Stat(paths.PlistPath); err == nil {
		err := runLaunchctl("bootout", paths.Target, paths.PlistPath)
		if err != nil && !isLaunchAgentNotLoaded(err) {
			return err
		}
	}
	fmt.Printf("Hometab service is not loaded.\n")
	return nil
}

func printLaunchAgentStatus(paths macServicePaths) error {
	if launchAgentLoaded(paths.ServiceName) {
		fmt.Printf("Hometab service is loaded.\n")
		fmt.Printf("URL: http://127.0.0.1:52173\n")
		fmt.Printf("Binary: %s\n", paths.BinPath)
		fmt.Printf("LaunchAgent: %s\n", paths.PlistPath)
		fmt.Printf("Logs: %s\n", paths.LogDir)
		return nil
	}
	fmt.Printf("Hometab service is not loaded.\n")
	fmt.Printf("LaunchAgent: %s\n", paths.PlistPath)
	return nil
}

func copyCurrentExecutable(dst string) error {
	src, err := os.Executable()
	if err != nil {
		return err
	}
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	if src == dst {
		return os.Chmod(dst, 0755)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func renderLaunchAgentPlist(paths macServicePaths) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>--no-open</string>
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
`, html.EscapeString(strings.TrimPrefix(paths.ServiceName, paths.Target+"/")), html.EscapeString(paths.BinPath), html.EscapeString(paths.StdoutPath), html.EscapeString(paths.StderrPath))
}

var errLaunchAgentNotLoaded = errors.New("launch agent is not loaded")

func isLaunchAgentNotLoaded(err error) bool {
	if errors.Is(err, errLaunchAgentNotLoaded) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "No such process") ||
		strings.Contains(msg, "Could not find specified service") ||
		strings.Contains(msg, "service is not loaded")
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
