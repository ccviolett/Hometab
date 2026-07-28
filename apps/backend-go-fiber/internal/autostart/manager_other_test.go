//go:build !darwin

package autostart

import (
	"errors"
	"testing"
)

func TestUnsupportedManager(t *testing.T) {
	manager := newManager()
	status, err := manager.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Supported || status.Platform == "" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if _, err := manager.Configure(Config{Enabled: true}); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Configure error = %v", err)
	}
}
