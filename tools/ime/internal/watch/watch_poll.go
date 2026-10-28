//go:build darwin || linux

package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/imectl"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/prefix"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/tmuxctl"
)

const (
	pollInterval    = 100 * time.Millisecond // IME state check (in-process, cheap)
	sessionInterval = 2 * time.Second        // look for new sessions and for server exit
	maxFailures     = 3                      // consecutive failed scans before giving up
)

// pollWait is the polling state source: it reads the IME state, then sleeps
// at most pollInterval.
func pollWait(d time.Duration) string {
	state, _ := imectl.State()
	time.Sleep(min(d, pollInterval))
	return state
}

// openLock opens the lock file next to the executable, which is private to
// the user who installed it; the shared temporary directory is only the
// fallback for an executable in a read-only place. All watchers of one
// executable therefore use the same file.
func openLock(name string) (*os.File, error) {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	dirs = append(dirs, os.TempDir())
	var err error
	for _, d := range dirs {
		var f *os.File
		if f, err = os.OpenFile(filepath.Join(d, name+".lock"), os.O_CREATE|os.O_RDWR, 0o600); err == nil {
			return f, nil
		}
	}
	return nil, err
}

// runWatch runs one resident watcher per socket. tmux runs `run-shell -b` once
// per server start, so a file lock keeps a leftover watcher of a previous
// server from running alongside the new one: the new watcher waits for the
// lock, and the old one exits when its server stops answering.
//
// newWait runs once the lock is held and returns the state source: wait blocks
// for at most d (less if the state changes) and returns the current IME state.
func runWatch(socket string, newWait func() func(d time.Duration) string) error {
	name := "ime-watch"
	if socket != "" {
		name += "-" + socket
	}
	lock, err := openLock(name)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock: %w", err)
	}

	defer prefix.Serve(socket)()
	wait := newWait()
	tmux := tmuxctl.Client{Socket: socket}
	var last string
	known := map[string]bool{}
	failures := 0
	nextAlive := time.Now().Add(sessionInterval)
	nextScan := time.Now().Add(sessionInterval)
	for {
		state := wait(min(time.Until(nextAlive), time.Until(nextScan)))
		now := time.Now()
		if state != last {
			// Notify tmux only when the mode changes.
			last = state
			known = push(tmux, state, known, true)
			nextScan = now.Add(sessionInterval)
		} else if now.After(nextScan) {
			// New sessions have no @ime yet; send the current mode to them only.
			known = push(tmux, state, known, false)
			nextScan = now.Add(sessionInterval)
		}
		if now.After(nextAlive) {
			// Stop once the server has stopped answering.
			nextAlive = now.Add(sessionInterval)
			if _, err := tmux.Sessions(); err == nil {
				failures = 0
			} else if failures++; failures >= maxFailures {
				return nil
			}
		}
	}
}
