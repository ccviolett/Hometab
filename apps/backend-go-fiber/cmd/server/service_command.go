package main

import (
	"fmt"
	"os"

	"hometab/internal/autostart"
)

type serviceCommandFlags struct {
	Install   bool
	Uninstall bool
	Start     bool
	Stop      bool
	Status    bool
}

func runServiceCommand(action string) error {
	manager := autostart.New()
	var err error
	switch action {
	case "install":
		err = manager.Install()
	case "uninstall":
		err = manager.Uninstall()
	case "start":
		err = manager.Start()
	case "stop":
		err = manager.Stop()
	case "status":
		// Status is printed below.
	default:
		return fmt.Errorf("unsupported service command: %s", action)
	}
	if err != nil {
		return err
	}

	status, err := manager.Status()
	if err != nil {
		return err
	}
	fmt.Printf("Platform: %s\n", status.Platform)
	fmt.Printf("Supported: %t\n", status.Supported)
	fmt.Printf("Login startup enabled: %t\n", status.Enabled)
	fmt.Printf("Service active: %t\n", status.Active)
	return nil
}

func handleServiceCommand(flags serviceCommandFlags) bool {
	action := ""
	count := 0
	if flags.Install {
		action = "install"
		count++
	}
	if flags.Uninstall {
		action = "uninstall"
		count++
	}
	if flags.Start {
		action = "start"
		count++
	}
	if flags.Stop {
		action = "stop"
		count++
	}
	if flags.Status {
		action = "status"
		count++
	}
	if count == 0 {
		return false
	}
	if count > 1 {
		fmt.Fprintln(os.Stderr, "only one service command can be used at a time")
		os.Exit(2)
	}
	if err := runServiceCommand(action); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", action, err)
		os.Exit(1)
	}
	return true
}
