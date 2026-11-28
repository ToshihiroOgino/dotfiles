// Command claude-tmux-state records Claude Code's state on the tmux pane so the
// window list can show it.
//
//	claude-tmux-state working|waiting|done|seen|clear [pane_id]
//
// States: working ◐ / waiting ○ / done ✓ (unchecked) / seen ● (checked)
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var icons = map[string]string{
	"working": "#[fg=colour136]◐#[fg=default]",
	"waiting": "#[fg=colour209]○#[fg=default]",
	"done":    "#[fg=colour71]✓#[fg=default]",
	"seen":    "#[fg=colour71]●#[fg=default]",
}

type paneStatus struct {
	visible bool
	state   string
}

func parseStatus(line string) paneStatus {
	fields := strings.Split(strings.TrimRight(line, "\r\n"), "|")
	if len(fields) != 4 {
		return paneStatus{}
	}
	attached := fields[2] != "" && fields[2] != "0"
	return paneStatus{
		visible: fields[0] == "1" && fields[1] == "1" && attached,
		state:   fields[3],
	}
}

func nextState(requested string, current paneStatus) (string, bool) {
	switch requested {
	// idle notifications arrive after Stop; only a running turn can be blocked
	case "waiting":
		return requested, current.state == "working"
	// already looking at the pane: nothing left to check
	case "done":
		if current.visible {
			return "seen", true
		}
	case "seen":
		return requested, current.state == "done"
	}
	return requested, true
}

func findTmux() (string, error) {
	names := []string{"tmux"}
	if runtime.GOOS == "windows" {
		names = []string{"psmux", "tmux"}
	}
	var err error
	for _, name := range names {
		var path string
		if path, err = exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", err
}

func run(requested, pane string) error {
	if _, ok := icons[requested]; !ok && requested != "clear" {
		return fmt.Errorf("unknown state %q", requested)
	}
	if pane == "" {
		return nil
	}
	tmux, err := findTmux()
	if err != nil {
		return nil
	}
	set := func(args ...string) error {
		return exec.Command(tmux, args...).Run()
	}

	if requested == "clear" {
		if err := set("set", "-pu", "-t", pane, "@claude"); err != nil {
			return err
		}
		return set("set", "-pu", "-t", pane, "@claude_icon")
	}

	out, err := exec.Command(tmux, "display", "-p", "-t", pane,
		"#{pane_active}|#{window_active}|#{session_attached}|#{@claude}").Output()
	if err != nil {
		return err
	}
	state, changed := nextState(requested, parseStatus(string(out)))
	if !changed {
		return nil
	}
	if err := set("set", "-p", "-t", pane, "@claude", state); err != nil {
		return err
	}
	return set("set", "-p", "-t", pane, "@claude_icon", icons[state])
}

func main() {
	var requested, pane string
	if len(os.Args) > 1 {
		requested = os.Args[1]
	}
	if len(os.Args) > 2 {
		pane = os.Args[2]
	}
	if pane == "" {
		pane = os.Getenv("TMUX_PANE")
	}
	if err := run(requested, pane); err != nil {
		fmt.Fprintln(os.Stderr, "claude-tmux-state:", err)
		os.Exit(1)
	}
}
