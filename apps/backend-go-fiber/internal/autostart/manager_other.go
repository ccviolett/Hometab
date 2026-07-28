//go:build !darwin

package autostart

import "runtime"

type unsupportedManager struct{}

func newManager() Manager {
	return unsupportedManager{}
}

func (unsupportedManager) Status() (Status, error) {
	return Status{Platform: runtime.GOOS, Supported: false}, nil
}

func (m unsupportedManager) Configure(Config) (Status, error) {
	status, _ := m.Status()
	return status, ErrUnsupported
}

func (unsupportedManager) Install() error   { return ErrUnsupported }
func (unsupportedManager) Uninstall() error { return ErrUnsupported }
func (unsupportedManager) Start() error     { return ErrUnsupported }
func (unsupportedManager) Stop() error      { return ErrUnsupported }
