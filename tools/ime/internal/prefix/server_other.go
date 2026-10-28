//go:build !darwin && !linux

package prefix

// Serve does nothing: there is no watcher to hand the sessions to.
func Serve(string) (stop func()) { return func() {} }

// Notify: the prefix session runs in the `ime prefix` process itself.
func Notify(socket, client string) bool { return false }
