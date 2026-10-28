package main

import "github.com/ToshihiroOgino/dotfiles/tools/ime/internal/tmuxctl"

// push sends state to the @ime option of every session (all) or only of
// sessions not in known, and returns the current session set. On error known
// is kept unchanged.
func push(c tmuxctl.Client, state string, known map[string]bool, all bool) map[string]bool {
	names, err := c.Sessions()
	if err != nil {
		return known
	}
	current := make(map[string]bool, len(names))
	for _, name := range names {
		if all || !known[name] {
			c.SetOption(name, "@ime", state)
		}
		current[name] = true
	}
	return current
}
