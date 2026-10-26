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

// Tracked is packages/<mgr>.json. Buckets is a pointer so that an empty list
// is written for scoop and the key is omitted for apt.
type Tracked struct {
	Comment  string    `json:"comment,omitempty"`
	Buckets  *[]string `json:"buckets,omitempty"`
	Packages []string  `json:"packages"`
	Ignore   []string  `json:"ignore"`
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
	if m.HasBuckets() && t.Buckets == nil {
		t.Buckets = &[]string{}
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
	AddPkgs, DelPkgs       []string
	AddBuckets, DelBuckets []string
}

func (d Diff) Empty() bool {
	return len(d.AddPkgs)+len(d.DelPkgs)+len(d.AddBuckets)+len(d.DelBuckets) == 0
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
	lines("+", green, "bucket:", d.AddBuckets)
	lines("-", red, "bucket:", d.DelBuckets)
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
	present := append(append([]string{}, cur...), keep(s.External)...)
	d.AddPkgs, d.DelPkgs = minus(cur, tr), minus(tr, present)
	if m.HasBuckets() {
		tb := keep(t.bucketList())
		s.Buckets = keep(s.Buckets)
		d.AddBuckets, d.DelBuckets = minus(s.Buckets, tb), minus(tb, s.Buckets)
	}
	return d, nil
}

func (t Tracked) bucketList() []string {
	if t.Buckets == nil {
		return nil
	}
	return *t.Buckets
}

// apply folds d into t. Removals are applied only when prune is set.
func apply(t *Tracked, d Diff, prune bool) {
	pkgs, buckets := d.AddPkgs, d.AddBuckets
	t.Packages = merge(t.Packages, pkgs)
	if t.Buckets != nil || len(buckets) > 0 {
		b := merge(t.bucketList(), buckets)
		if prune {
			b = minus(b, d.DelBuckets)
		}
		t.Buckets = &b
	}
	if prune {
		t.Packages = minus(t.Packages, d.DelPkgs)
	}
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
