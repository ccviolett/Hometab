package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenBrowserWhenServerIsReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var openedURL string
	openBrowserWhenReadyWith(server.URL, func(targetURL string) error {
		openedURL = targetURL
		return nil
	})

	assert.Equal(t, server.URL, openedURL)
}

func TestBrowserCommand(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		wantName string
		wantArgs []string
	}{
		{name: "macOS", goos: "darwin", wantName: "open", wantArgs: []string{"http://127.0.0.1:52173"}},
		{name: "Windows", goos: "windows", wantName: "rundll32", wantArgs: []string{"url.dll,FileProtocolHandler", "http://127.0.0.1:52173"}},
		{name: "Linux", goos: "linux", wantName: "xdg-open", wantArgs: []string{"http://127.0.0.1:52173"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, err := browserCommand(tt.goos, "http://127.0.0.1:52173")
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestBrowserCommandRejectsUnsupportedPlatform(t *testing.T) {
	name, args, err := browserCommand("plan9", "http://127.0.0.1:52173")

	require.Error(t, err)
	assert.Empty(t, name)
	assert.Nil(t, args)
}
