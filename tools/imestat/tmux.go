package main

import (
	"os/exec"
	"strings"
)

// tmuxClient runs tmux commands against one socket.
type tmuxClient struct {
	socket string
}

func (c tmuxClient) run(args ...string) (string, error) {
	if c.socket != "" {
		args = append([]string{"-L", c.socket}, args...)
	}
	cmd := exec.Command("tmux", args...)
	hideConsole(cmd)
	out, err := cmd.Output()
	return string(out), err
}

// sessions lists session names. psmux routes -t by name, not by session_id.
func (c tmuxClient) sessions() ([]string, error) {
	out, err := c.run("list-sessions", "-F", "#{session_name}")
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

// setIME stores state in @ime. -t is required: psmux runs one server per
// session, so a bare `set -g` reaches only one of them.
func (c tmuxClient) setIME(session, state string) {
	c.run("set-option", "-t", session, "-g", "@ime", state)
}

// push sends state to every session (all) or only to sessions not in known,
// and returns the current session set. On error known is kept unchanged.
func (c tmuxClient) push(state string, known map[string]bool, all bool) map[string]bool {
	names, err := c.sessions()
	if err != nil {
		return known
	}
	current := make(map[string]bool, len(names))
	for _, name := range names {
		if all || !known[name] {
			c.setIME(name, state)
		}
		current[name] = true
	}
	return current
}
