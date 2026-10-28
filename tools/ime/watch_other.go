//go:build !windows && !darwin && !linux

package main

import (
	"fmt"
	"runtime"
)

func watch(string, uint32) error {
	return fmt.Errorf("watch not supported on %s", runtime.GOOS)
}
