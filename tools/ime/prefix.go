package main

import (
	"time"

	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/imectl"
)

const (
	prefixPoll    = 50 * time.Millisecond
	prefixEnter   = time.Second      // how long to wait for the client to reach the prefix table
	prefixMaxHold = 10 * time.Minute // give up waiting for the prefix sequence to finish
)

// prefixSession switches the IME to alphanumeric input while client sits in
// tmux's prefix key table, then restores the previous input source. It is
// meant to be started by the C-b binding next to `switch-client -T prefix`, so
// the prefix table may not be entered yet when it starts.
//
// It never fails: run-shell would show any output or non-zero exit in the
// pane on every prefix press, which is worse than a missing IME switch (e.g.
// when no IME tool is installed).
func prefixSession(client string, table func(client string) (string, error)) {
	prev, err := imectl.Save()
	if err != nil {
		return
	}
	if err := imectl.Set(false); err != nil {
		return
	}
	cur, err := imectl.Save()
	if err != nil || cur == prev {
		return // already alphanumeric (or unreadable): nothing to restore
	}
	defer func() {
		// Leave it alone if the user switched the IME themselves meanwhile.
		if now, err := imectl.Save(); err == nil && now == cur {
			imectl.Restore(prev)
		}
	}()

	// The binding starts us in parallel with switch-client, so wait for the
	// prefix table to show up before waiting for it to end.
	for deadline := time.Now().Add(prefixEnter); time.Now().Before(deadline); time.Sleep(prefixPoll) {
		t, err := table(client)
		if err != nil {
			return
		}
		if t == "prefix" {
			break
		}
	}
	for deadline := time.Now().Add(prefixMaxHold); time.Now().Before(deadline); time.Sleep(prefixPoll) {
		if t, err := table(client); err != nil || t != "prefix" {
			return
		}
	}
}
