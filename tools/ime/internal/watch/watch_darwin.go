//go:build darwin

package watch

import "time"

// Run polls the input source; macOS has no change notification used here.
// serverPID is only used on Windows and ignored here.
func Run(socket string, _ uint32) error {
	return runWatch(socket, func() func(time.Duration) string { return pollWait })
}
