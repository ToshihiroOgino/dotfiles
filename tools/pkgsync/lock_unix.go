//go:build unix

package main

import (
	"errors"
	"os"
	"syscall"
)

// lockFile takes a non-blocking exclusive flock, or returns errLocked.
func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return errLocked
	}
	return err
}
