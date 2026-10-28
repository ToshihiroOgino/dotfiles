package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	checkInterval = 5 * time.Minute
	maxLogs       = 50 // older failure logs are deleted beyond this
)

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

// cmdDiff prints the difference and refreshes packages/pending/. With
// throttle (shell startup) it does nothing while the last successful run is
// less than checkInterval old. A failed run writes a log under log/pkgsync/
// and leaves the stamp alone, so that it is retried and reported.
func cmdDiff(root string, throttle bool) error {
	ms := detect()
	if len(ms) == 0 {
		return errNoManager
	}
	stamp := filepath.Join(root, "packages", "pending", ".stamp")
	// Checked before the lock so that most shell startups never contend for
	// it, and again after, in case a run finished while this one waited.
	fresh := func() bool {
		fi, err := os.Stat(stamp)
		return err == nil && time.Since(fi.ModTime()) < checkInterval
	}
	if throttle && fresh() {
		return nil
	}
	release, err := acquire(root)
	if err != nil {
		return err
	}
	defer release()
	if throttle {
		if fresh() {
			return nil
		}
	} else {
		reportLastFailure(root, stamp)
	}
	p := painter(colorEnabled(os.Stdout))
	var errs []error
	for _, m := range ms {
		_, s, d, err := state(root, m)
		if err != nil {
			errs = append(errs, err)
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
		if err := writePending(root, m, d, s.Warnings); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		if lerr := writeFailureLog(root, err); lerr != nil {
			err = errors.Join(err, lerr)
		}
		return err
	}
	return os.WriteFile(stamp, []byte(time.Now().Format(time.RFC3339)), 0o644)
}

func logDir(root string) string { return filepath.Join(root, "log", "pkgsync") }

// writeFailureLog writes cause to log/pkgsync/<UTC time>.log. A name that is
// already taken gets -1, -2, ... appended.
func writeFailureLog(root string, cause error) error {
	dir := logDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	for i := 0; ; i++ {
		name := ts + ".log"
		if i > 0 {
			name = fmt.Sprintf("%s-%d.log", ts, i)
		}
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(f, cause)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		pruneLogs(dir)
		return err
	}
}

// sortedLogs returns the paths of the logs in dir, oldest first.
func sortedLogs(dir string) []string {
	type entry struct {
		path, ts string
		seq      int
	}
	ents, _ := os.ReadDir(dir)
	var logs []entry
	for _, e := range ents {
		base, ok := strings.CutSuffix(e.Name(), ".log")
		if !ok {
			continue
		}
		ts, n, _ := strings.Cut(base, "-")
		seq, _ := strconv.Atoi(n)
		logs = append(logs, entry{filepath.Join(dir, e.Name()), ts, seq})
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].ts != logs[j].ts {
			return logs[i].ts < logs[j].ts
		}
		return logs[i].seq < logs[j].seq
	})
	paths := make([]string, len(logs))
	for i, l := range logs {
		paths[i] = l.path
	}
	return paths
}

// latestLog returns the path of the newest log in dir, or "" if there is none.
func latestLog(dir string) string {
	logs := sortedLogs(dir)
	if len(logs) == 0 {
		return ""
	}
	return logs[len(logs)-1]
}

// pruneLogs deletes the oldest logs beyond maxLogs.
func pruneLogs(dir string) {
	logs := sortedLogs(dir)
	for len(logs) > maxLogs {
		os.Remove(logs[0])
		logs = logs[1:]
	}
}

// reportLastFailure prints the newest log if it is newer than the stamp, that
// is, if no run has succeeded since it was written.
func reportLastFailure(root, stamp string) {
	path := latestLog(logDir(root))
	if path == "" {
		return
	}
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if st, err := os.Stat(stamp); err == nil && !fi.ModTime().After(st.ModTime()) {
		return
	}
	if b, err := os.ReadFile(path); err == nil {
		fmt.Fprintf(os.Stderr, "the last run failed (%s):\n%s\n", path, b)
	}
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

// cmdApply writes the additions and removals to packages/<mgr>.json and
// clears the manager's pending diff and warnings, so the prompt updates at
// once; the next diff brings the warnings back if they still apply.
func cmdApply(root string) error {
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
	p := painter(colorEnabled(os.Stdout))
	var errs []error
	var changed bool
	for _, m := range ms {
		t, _, d, err := state(root, m)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fmt.Println(p.paint(bold, "["+m.Name()+"]"))
		if d.Empty() {
			fmt.Println("nothing to apply")
		} else {
			apply(&t, d)
			if err := saveTracked(root, m, t); err != nil {
				errs = append(errs, err)
				continue
			}
			fmt.Print(d.format(p))
			changed = true
		}
		if err := clearPending(root, m); err != nil {
			errs = append(errs, err)
		}
	}
	if changed {
		fmt.Println("please commit and push packages/")
	}
	return errors.Join(errs...)
}

func clearPending(root string, m Manager) error {
	dir := filepath.Join(root, "packages", "pending")
	for _, ext := range []string{".diff", ".warn"} {
		if err := putFile(filepath.Join(dir, m.Name()+ext), ""); err != nil {
			return err
		}
	}
	return updateSummary(dir)
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
