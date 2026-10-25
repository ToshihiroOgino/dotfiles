//go:build !windows && !darwin && !linux

package main

import (
	"fmt"
	"runtime"
)

func imeState() (string, error) {
	return "", fmt.Errorf("not supported on %s", runtime.GOOS)
}
