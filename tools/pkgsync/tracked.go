package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Tracked is packages/<mgr>.json. Repos are the extra package sources:
// buckets (scoop), taps (brew), PPAs (apt).
type Tracked struct {
	Comment  string   `json:"comment,omitempty"`
	Repos    []string `json:"repos"`
	Packages []string `json:"packages"`
	Ignore   []string `json:"ignore"`
}

func trackedPath(root string, m Manager) string {
	return filepath.Join(root, "packages", m.Name()+".json")
}

func loadTracked(root string, m Manager) (Tracked, error) {
	var t Tracked
	b, err := os.ReadFile(trackedPath(root, m))
	if errors.Is(err, fs.ErrNotExist) {
		return t, nil
	}
	if err != nil {
		return t, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return t, fmt.Errorf("%s: %w", trackedPath(root, m), err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return t, fmt.Errorf("%s: unexpected data after the JSON object", trackedPath(root, m))
	}
	return t, nil
}

func saveTracked(root string, m Manager, t Tracked) error {
	if t.Repos == nil {
		t.Repos = []string{}
	}
	if t.Packages == nil {
		t.Packages = []string{}
	}
	if t.Ignore == nil {
		t.Ignore = []string{}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(t); err != nil {
		return err
	}
	p := trackedPath(root, m)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return writeFile(p, buf.Bytes())
}

// Diff is the difference between this machine and the tracked list.
type Diff struct {
	AddPkgs, DelPkgs   []string
	AddRepos, DelRepos []string
}

func (d Diff) Empty() bool {
	return len(d.AddPkgs)+len(d.DelPkgs)+len(d.AddRepos)+len(d.DelRepos) == 0
}

// String is the plain form written to packages/pending/*.diff.
func (d Diff) String() string { return d.format(false) }

func (d Diff) format(p painter) string {
	var sb strings.Builder
	lines := func(sign, code, prefix string, xs []string) {
		for _, x := range xs {
			fmt.Fprintln(&sb, p.paint(code, sign+" "+prefix+x))
		}
	}
	lines("+", green, "repo:", d.AddRepos)
	lines("-", red, "repo:", d.DelRepos)
	lines("+", green, "", d.AddPkgs)
	lines("-", red, "", d.DelPkgs)
	return sb.String()
}

func compare(m Manager, t Tracked, s Snapshot) (Diff, error) {
	ign, err := parseIgnore(t.Ignore)
	if err != nil {
		return Diff{}, err
	}
	keep := func(xs []string) []string {
		var out []string
		for _, x := range xs {
			if !ign.Match(x) {
				out = append(out, x)
			}
		}
		return out
	}

	var d Diff
	cur, tr := keep(s.Packages), keep(t.Packages)
	// A tracked package that is installed from outside the repositories (a
	// local .deb, a URL manifest) is still installed: not a removal.
	// Likewise for packages the OS installer put there.
	present := append(append([]string{}, cur...), keep(s.External)...)
	present = append(present, keep(s.Initial)...)
	// Auto-installed packages are not reported as manual, but are present.
	present = append(present, keep(s.Installed)...)
	d.AddPkgs, d.DelPkgs = minus(cur, tr), minus(tr, present)
	curRepos, trRepos := keep(s.Repos), keep(t.Repos)
	d.AddRepos, d.DelRepos = minus(curRepos, trRepos), minus(trRepos, curRepos)
	return d, nil
}

// apply folds d into t: additions and removals.
func apply(t *Tracked, d Diff) {
	t.Packages = minus(merge(t.Packages, d.AddPkgs), d.DelPkgs)
	t.Repos = minus(merge(t.Repos, d.AddRepos), d.DelRepos)
}

// minus returns the sorted elements of a that are not in b.
func minus(a, b []string) []string {
	skip := make(map[string]bool, len(b))
	for _, x := range b {
		skip[x] = true
	}
	var out []string
	for _, x := range a {
		if !skip[x] {
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

// merge returns the sorted union of a and b without duplicates.
func merge(a, b []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, x := range append(append([]string{}, a...), b...) {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}
