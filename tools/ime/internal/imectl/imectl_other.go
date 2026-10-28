//go:build !windows && !darwin && !linux

package imectl

import (
	"fmt"
	"runtime"
)

func State() (string, error) {
	return "", fmt.Errorf("not supported on %s", runtime.GOOS)
}

func Set(bool) error {
	return fmt.Errorf("not supported on %s", runtime.GOOS)
}

func Save() (string, error) {
	return "", fmt.Errorf("not supported on %s", runtime.GOOS)
}

func Restore(string) error {
	return fmt.Errorf("not supported on %s", runtime.GOOS)
}
