package main

import (
	"fmt"
	"path"
	"strings"
)

// Ignore is a list of gitignore-style patterns for package names.
//
// Supported: "*", "?", "[...]" (negate with "[^...]"), "\" escapes, "!" to
// re-include, "#" comments and blank lines. The last matching line wins.
// A name is also ignored when a pattern matches a leading "/"-separated part
// of it, as in gitignore: "hashicorp/tap" ignores "hashicorp/tap/terraform".
// Trailing-slash directory rules and "**" do not exist.
type Ignore struct {
	rules []ignoreRule
}

type ignoreRule struct {
	pattern string
	negate  bool
}

func parseIgnore(lines []string) (Ignore, error) {
	var g Ignore
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		r := ignoreRule{pattern: l}
		if strings.HasPrefix(l, "!") {
			r.negate, r.pattern = true, l[1:]
		}
		if _, err := path.Match(r.pattern, ""); err != nil {
			return Ignore{}, fmt.Errorf("ignore %q: %w", l, err)
		}
		g.rules = append(g.rules, r)
	}
	return g, nil
}

func (g Ignore) Match(name string) bool {
	ignored := false
	for _, r := range g.rules {
		if matchPrefix(r.pattern, name) {
			ignored = !r.negate
		}
	}
	return ignored
}

// matchPrefix reports whether pattern matches name or any leading part of it
// that ends before a "/".
func matchPrefix(pattern, name string) bool {
	for {
		if ok, _ := path.Match(pattern, name); ok {
			return true
		}
		i := strings.LastIndex(name, "/")
		if i < 0 {
			return false
		}
		name = name[:i]
	}
}
