package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	prefixPoll    = 50 * time.Millisecond
	prefixEnter   = time.Second      // how long to wait for the client to reach the prefix table
	prefixMaxHold = 10 * time.Minute // give up waiting for the prefix sequence to finish
)

// clientFormat is the list-clients format that clientTable parses.
const clientFormat = "#{client_name}\t#{client_key_table}"

// tableFunc reports the key table that a tmux client is in.
type tableFunc func(client string) (string, error)

// spawnTable looks the table up by running tmux for every call.
func spawnTable(tc tmuxClient) tableFunc {
	return func(client string) (string, error) {
		out, err := tc.run("list-clients", "-F", clientFormat)
		if err != nil {
			return "", err
		}
		return tableOf(out, client)
	}
}

func tableOf(out, client string) (string, error) {
	if t, ok := clientTable(out, client); ok {
		return t, nil
	}
	return "", fmt.Errorf("tmux client %q not found", client)
}

// prefixSession switches the IME to alphanumeric input while client sits in
// tmux's prefix key table, then restores the previous input source. It is
// meant to be started by the C-b binding next to `switch-client -T prefix`, so
// the prefix table may not be entered yet when it starts.
//
// It never fails: run-shell would show any output or non-zero exit in the
// pane on every prefix press, which is worse than a missing IME switch (e.g.
// when no IME tool is installed).
func prefixSession(client string, table func(client string) (string, error)) {
	prev, err := saveIME()
	if err != nil {
		return
	}
	if err := setIME(false); err != nil {
		return
	}
	cur, err := saveIME()
	if err != nil || cur == prev {
		return // already alphanumeric (or unreadable): nothing to restore
	}
	defer func() {
		// Leave it alone if the user switched the IME themselves meanwhile.
		if now, err := saveIME(); err == nil && now == cur {
			restoreIME(prev)
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

// clientTable finds client's key table in `list-clients -F
// "#{client_name}\t#{client_key_table}"` output.
func clientTable(out, client string) (string, bool) {
	for _, line := range strings.Split(out, "\n") {
		if name, table, ok := strings.Cut(strings.TrimSpace(line), "\t"); ok && name == client {
			return table, true
		}
	}
	return "", false
}
