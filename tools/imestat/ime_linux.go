//go:build linux

package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

const queryTimeout = 500 * time.Millisecond

// imeState maps the current fcitx5 input method name to a mode. Plain
// keyboard layouts are named keyboard-<layout> (keyboard-us, keyboard-jp,
// keyboard-us-dvorak, ...) whatever the physical layout is, so those mean
// alphanumeric input and any other name is an input method engine (mozc, ...).
func imeState() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "fcitx5-remote", "-n").Output()
	if errors.Is(err, exec.ErrNotFound) {
		return "", err // not installed: fail so `if-shell 'imestat'` skips the watcher
	}
	if err != nil {
		return "?", nil // fcitx5 not running
	}
	return modeOf(strings.TrimSpace(string(out))), nil
}

// modeOf maps an input method name to "あ", "A" or "?".
func modeOf(name string) string {
	switch {
	case name == "":
		return "?"
	case strings.HasPrefix(name, "keyboard-"):
		return "A"
	default:
		return "あ"
	}
}
