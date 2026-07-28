package autostart

import "errors"

var ErrUnsupported = errors.New("login startup is not supported on this platform")

type Status struct {
	Platform  string `json:"platform"`
	Supported bool   `json:"supported"`
	Enabled   bool   `json:"enabled"`
	Active    bool   `json:"active"`
}

type Config struct {
	Enabled bool `json:"enabled"`
}

type Manager interface {
	Status() (Status, error)
	Configure(Config) (Status, error)
	Install() error
	Uninstall() error
	Start() error
	Stop() error
}

func New() Manager {
	return newManager()
}
