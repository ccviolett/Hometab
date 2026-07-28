package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Config struct {
	Server           ServerConfig          `mapstructure:"server"`
	Database         DatabaseConfig        `mapstructure:"database"`
	Log              LogConfig             `mapstructure:"log"`
	ExternalRequests ExternalRequestConfig `mapstructure:"external_requests"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type ExternalRequestConfig struct {
	AllowRemoteExecution bool `mapstructure:"allow_remote_execution"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Pretty bool   `mapstructure:"pretty"`
}

func (c Config) LogLevel() zerolog.Level {
	switch strings.ToLower(c.Log.Level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func Load() *Config {
	// Sensible defaults: local-only, high port, works without config.yaml.
	viper.SetDefault("server.port", 52173)
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("database.path", defaultDatabasePath())
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.pretty", true)
	viper.SetDefault("external_requests.allow_remote_execution", false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")

	viper.SetEnvPrefix("HOME")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Config file is optional — defaults + env vars are sufficient
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Sprintf("failed to read config: %v", err))
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
	if strings.TrimSpace(cfg.Database.Path) == "" {
		cfg.Database.Path = defaultDatabasePath()
	}

	return &cfg
}

func defaultDatabasePath() string {
	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		legacyPath := filepath.Join(configDir, "Home", "data.db")
		if fileExists(legacyPath) {
			return legacyPath
		}
		return filepath.Join(configDir, "Hometab", "data.db")
	}

	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		legacyPath := filepath.Join(homeDir, ".home", "data.db")
		if fileExists(legacyPath) {
			return legacyPath
		}
		return filepath.Join(homeDir, ".hometab", "data.db")
	}

	return "./data.db"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
