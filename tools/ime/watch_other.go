//go:build !windows && !darwin && !linux

package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func watch(string, uint32) error {
	return fmt.Errorf("watch not supported on %s", runtime.GOOS)
}

func hideConsole(*exec.Cmd) {}
