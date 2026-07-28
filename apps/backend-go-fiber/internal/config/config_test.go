package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"WARN", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"ERROR", zerolog.ErrorLevel},
		{"unknown", zerolog.InfoLevel},
		{"", zerolog.InfoLevel},
	}
	for _, tt := range tests {
		c := Config{Log: LogConfig{Level: tt.input}}
		assert.Equal(t, tt.expected, c.LogLevel(), "input: %q", tt.input)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Load()
	assert.Equal(t, 52173, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, defaultDatabasePath(), cfg.Database.Path)
	assert.Equal(t, true, cfg.Log.Pretty)
}

func TestDefaultDatabasePathUsesHometabDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := defaultDatabasePath()

	assert.Contains(t, path, "Hometab")
	assert.Equal(t, "data.db", filepath.Base(path))
}

func TestDefaultDatabasePathFindsExistingHomeDatabase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir, err := os.UserConfigDir()
	assert.NoError(t, err)
	legacyPath := filepath.Join(configDir, "Home", "data.db")
	assert.NoError(t, os.MkdirAll(filepath.Dir(legacyPath), 0o755))
	assert.NoError(t, os.WriteFile(legacyPath, []byte("existing"), 0o600))

	assert.Equal(t, legacyPath, defaultDatabasePath())
}
