//go:build !darwin && !linux

package main

// notifyDaemon: the prefix session runs in the `ime prefix` process itself.
func notifyDaemon(socket, client string) bool { return false }
