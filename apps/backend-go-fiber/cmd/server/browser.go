package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

const browserReadyTimeout = 15 * time.Second

func openBrowserWhenReady(targetURL string) {
	openBrowserWhenReadyWith(targetURL, openBrowser)
}

func openBrowserWhenReadyWith(targetURL string, opener func(string) error) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	healthURL := targetURL + "/health"
	deadline := time.Now().Add(browserReadyTimeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if err := opener(targetURL); err != nil {
					log.Warn().Err(err).Str("url", targetURL).Msg("Unable to open the default browser")
				}
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Warn().Str("url", targetURL).Msg("Browser was not opened because the server did not become ready")
}

func openBrowser(targetURL string) error {
	name, args, err := browserCommand(runtime.GOOS, targetURL)
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func browserCommand(goos, targetURL string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "open", []string{targetURL}, nil
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", targetURL}, nil
	case "linux":
		return "xdg-open", []string{targetURL}, nil
	default:
		return "", nil, fmt.Errorf("opening a browser is not supported on %s", goos)
	}
}
