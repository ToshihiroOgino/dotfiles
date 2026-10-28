//go:build !darwin && !linux

package prefix

// Notify: the prefix session runs in the `ime prefix` process itself.
func Notify(socket, client string) bool { return false }
