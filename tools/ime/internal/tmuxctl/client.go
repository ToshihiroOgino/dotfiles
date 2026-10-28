// Package tmuxctl runs tmux commands: one process per call (Client), or over
// a single control-mode connection (Pool).
package tmuxctl

import (
	"os/exec"
	"strings"
)

// Client runs tmux commands against one socket.
type Client struct {
	Socket string // tmux socket name for -L; the default socket if empty
}

// Run runs tmux with args and returns its standard output.
func (c Client) Run(args ...string) (string, error) {
	if c.Socket != "" {
		args = append([]string{"-L", c.Socket}, args...)
	}
	cmd := exec.Command("tmux", args...)
	hideConsole(cmd)
	out, err := cmd.Output()
	return string(out), err
}

// Sessions lists session names. psmux routes -t by name, not by session_id.
func (c Client) Sessions() ([]string, error) {
	out, err := c.Run("list-sessions", "-F", "#{session_name}")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if name := strings.TrimSpace(line); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

// SetOption sets a global option of one session. -t is required: psmux runs
// one server per session, so a bare `set -g` reaches only one of them.
func (c Client) SetOption(session, name, value string) {
	c.Run("set-option", "-t", session, "-g", name, value)
}
