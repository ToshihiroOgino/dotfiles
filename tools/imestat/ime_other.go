//go:build !windows

package main

import (
	"fmt"
	"runtime"
)

func imeState() (string, error) {
	return "", fmt.Errorf("not supported on %s", runtime.GOOS)
}
