//go:build !darwin

package main

import "fmt"

func runServiceCommand(action string) error {
	return fmt.Errorf("%s is only supported on macOS for now", action)
}
