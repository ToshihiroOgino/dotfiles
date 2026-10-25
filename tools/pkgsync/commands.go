package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const checkInterval = 5 * time.Minute

// state loads the tracked list and the machine snapshot and compares them.
func state(root string, m Manager) (Tracked, Snapshot, Diff, error) {
	t, err := loadTracked(root, m)
	if err != nil {
		return t, Snapshot{}, Diff{}, err
	}
	s, err := m.Snapshot()
	if err != nil {
		return t, s, Diff{}, fmt.Errorf("%s: %w", m.Name(), err)
	}
	d, err := compare(m, t, s)
	return t, s, d, err
}

func cmdDiff(root string) error {
	ms := detect()
	if len(ms) == 0 {
		return errNoManager
	}
	// A failed background check leaves its error here; nothing else shows it.
	if b, err := os.ReadFile(filepath.Join(root, ".local", "pkgsync.err")); err == nil {
		fmt.Fprintf(os.Stderr, "the last background check failed:\n%s\n", b)
	}
	p := painter(colorEnabled(os.Stdout))
	var failed bool
	for _, m := range ms {
		_, s, d, err := state(root, m)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pkgsync:", err)
			failed = true
			continue
		}
		fmt.Println(p.paint(bold, "["+m.Name()+"]"))
		if d.Empty() {
			fmt.Println("no difference")
		} else {
			fmt.Print(d.format(p))
		}
		for _, w := range s.Warnings {
			fmt.Println(p.paint(yellow, "warning: "+w))
		}
		for _, h := range m.InstallHints(d) {
			fmt.Println(p.paint(cyan, "install: "+h))
		}
		if len(s.External) > 0 {
			fmt.Println("outside repositories:", strings.Join(s.External, ", "))
		}
	}
	if failed {
		return errors.New("some managers failed")
	}
	return nil
}

// cmdInit creates an empty packages/<mgr>.json for every supported manager
// that has none, installed here or not. Existing files are left alone.
func cmdInit(root string) error {
	release, err := acquire(root)
	if err != nil {
		return err
	}
	defer release()
	for _, m := range allManagers {
		p := trackedPath(root, m)
		if _, err := os.Stat(p); err == nil {
			fmt.Println("exists  ", p)
			continue
		}
		if err := saveTracked(root, m, Tracked{}); err != nil {
			return err
		}
		fmt.Println("created ", p)
	}
	return nil
}

func cmdApply(root string, prune bool) error {
	ms := detect()
	if len(ms) == 0 {
		return errNoManager
	}
	release, err := acquire(root)
	if err != nil {
		return err
	}
	defer release()
	// A failure in one manager must not stop the others or hide the
	// "commit and push" reminder for what was already written.
	var errs []error
	var changed bool
	for _, m := range ms {
		t, s, d, err := state(root, m)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fmt.Printf("[%s]\n", m.Name())
		// Removals that add leaves alone stay pending.
		var rest Diff
		if !prune {
			rest = Diff{DelPkgs: d.DelPkgs, DelBuckets: d.DelBuckets}
			d.DelPkgs, d.DelBuckets = nil, nil
		}
		if d.Empty() {
			fmt.Println("nothing to apply")
		} else {
			apply(&t, d, prune)
			if err := saveTracked(root, m, t); err != nil {
				errs = append(errs, err)
				continue
			}
			fmt.Print(d)
			changed = true
		}
		if err := writePending(root, m, rest, s.Warnings); err != nil {
			errs = append(errs, err)
		}
	}
	if changed {
		fmt.Println("please commit and push packages/")
	}
	return errors.Join(errs...)
}

// cmdCheck is run in the background at shell startup, so it never fails
// loudly: errors go to .local/pkgsync.err, which `pkgsync diff` prints.
func cmdCheck(root string) error {
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	release, err := acquire(root)
	if err != nil {
		return nil // locked: another shell is already checking
	}
	defer release()

	dir := filepath.Join(root, ".local")
	stamp := filepath.Join(dir, "pkgsync.stamp")
	if fi, err := os.Stat(stamp); err == nil && time.Since(fi.ModTime()) < checkInterval {
		return nil
	}
	if os.WriteFile(stamp, []byte(time.Now().Format(time.RFC3339)), 0o644) != nil {
		return nil
	}

	var errs []error
	for _, m := range detect() {
		if err := refreshPending(root, m); err != nil {
			errs = append(errs, err)
		}
	}
	errFile := filepath.Join(dir, "pkgsync.err")
	if err := errors.Join(errs...); err != nil {
		os.WriteFile(errFile, []byte(err.Error()+"\n"), 0o644)
	} else {
		os.Remove(errFile)
	}
	return nil
}

// refreshPending writes packages/pending/<mgr>.diff and <mgr>.warn, or
// removes them when there is nothing to report, then refreshes the summary.
func refreshPending(root string, m Manager) error {
	_, s, d, err := state(root, m)
	if err != nil {
		return err
	}
	return writePending(root, m, d, s.Warnings)
}

func writePending(root string, m Manager, d Diff, warnings []string) error {
	dir := filepath.Join(root, "packages", "pending")
	warn := strings.Join(warnings, "\n")
	if warn != "" {
		warn += "\n"
	}
	if err := putFile(filepath.Join(dir, m.Name()+".diff"), d.String()); err != nil {
		return err
	}
	if err := putFile(filepath.Join(dir, m.Name()+".warn"), warn); err != nil {
		return err
	}
	return updateSummary(dir)
}
