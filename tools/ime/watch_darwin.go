//go:build darwin

package main

import "time"

// watch polls the input source; macOS has no change notification used here.
// serverPID is only used on Windows and ignored here.
func watch(socket string, _ uint32) error {
	return runWatch(socket, func() func(time.Duration) string { return pollWait })
}
