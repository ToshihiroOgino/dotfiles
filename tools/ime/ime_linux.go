//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
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
		return "", err // not installed: fail so `if-shell 'ime'` skips the watcher
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

// setIME deactivates fcitx5, which falls back to the group's first input
// method (the plain keyboard layout), or activates it again.
func setIME(native bool) error {
	flag := "-c"
	if native {
		flag = "-o"
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	if err := exec.CommandContext(ctx, "fcitx5-remote", flag).Run(); err != nil {
		return fmt.Errorf("fcitx5-remote %s: %w", flag, err)
	}
	return nil
}

// saveIME returns the current input method name.
func saveIME() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "fcitx5-remote", "-n").Output()
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "", fmt.Errorf("no input method")
	}
	return name, nil
}

// restoreIME switches back to the input method a saveIME token names.
func restoreIME(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	if err := exec.CommandContext(ctx, "fcitx5-remote", "-s", token).Run(); err != nil {
		return fmt.Errorf("fcitx5-remote -s %s: %w", token, err)
	}
	return nil
}
