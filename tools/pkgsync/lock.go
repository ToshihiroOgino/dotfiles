package main

import (
	"errors"
	"os"
	"path/filepath"
)

var errLocked = errors.New("another pkgsync is running")

// acquire takes an exclusive OS lock on <root>/.local/pkgsync.lock. It never
// waits: of several simultaneous callers exactly one succeeds and the rest
// get errLocked at once. The OS drops the lock when the holder exits, even
// if it crashes, so there is no stale lock to clean up. The file itself is
// left in place on purpose; removing it would reopen a race.
func acquire(root string) (release func(), err error) {
	dir := filepath.Join(root, ".local")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, "pkgsync.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return func() { f.Close() }, nil
}

// writeFile replaces path atomically, so a reader never sees a half-written
// file. The temporary name does not end in .diff or .json.
func writeFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), "*.tmp")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(f.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}
