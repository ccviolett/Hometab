package main

import (
	"fmt"
	"os"
)

type serviceCommandFlags struct {
	Install   bool
	Uninstall bool
	Start     bool
	Stop      bool
	Status    bool
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
