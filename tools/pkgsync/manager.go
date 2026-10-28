package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

// Snapshot is what a package manager reports for this machine.
type Snapshot struct {
	Packages  []string
	Buckets   []string // only for managers with HasBuckets
	External  []string // installed from outside the repositories; not tracked
	Initial   []string // installed by the OS installer; not tracked, but present
	Installed []string // every installed package, auto-installed ones included; only for presence
	Warnings  []string // things to look at; only scoop reports any (Info)
}

type Manager interface {
	Name() string
	Available() bool
	HasBuckets() bool
	Snapshot() (Snapshot, error)
	// InstallHints returns commands that would install what d.DelPkgs and
	// d.DelBuckets list: tracked, but missing here. They are only printed.
	InstallHints(d Diff) []string
}

// detect and output are variables so that tests can substitute them.
var allManagers = []Manager{apt{}, scoop{}, brew{}}

var detect = func() []Manager {
	var ms []Manager
	for _, m := range allManagers {
		if m.Available() {
			ms = append(ms, m)
		}
	}
	return ms
}

var output = func(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	return cmd.Output()
}

func available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

var errNoManager = errors.New("no supported package manager found")

func lines(b []byte) []string {
	var out []string
	for _, l := range bytes.Split(b, []byte("\n")) {
		if s := string(bytes.TrimSpace(l)); s != "" {
			out = append(out, s)
		}
	}
	return out
}
