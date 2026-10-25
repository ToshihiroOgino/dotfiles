package main

import (
	"strings"
	"testing"
)

func TestIgnoreMatch(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		pkg   string
		want  bool
	}{
		{"empty list", nil, "git", false},
		{"exact", []string{"git"}, "git", true},
		{"exact is not a prefix match", []string{"vim"}, "neovim", false},
		{"exact is not a suffix match", []string{"vim"}, "vim-tiny", false},
		{"star suffix", []string{"ubuntu-*"}, "ubuntu-desktop", true},
		{"star suffix needs the prefix", []string{"ubuntu-*"}, "xubuntu-desktop", false},
		{"star matches empty", []string{"lib*"}, "lib", true},
		{"question mark", []string{"python3.?"}, "python3.9", true},
		{"question mark needs one char", []string{"python3.?"}, "python3.", false},
		{"question mark does not match two", []string{"python3.?"}, "python3.10", false},
		{"class", []string{"gcc-[0-9]*"}, "gcc-12", true},
		{"class mismatch", []string{"gcc-[0-9]*"}, "gcc-doc", false},
		{"negated class", []string{"[^a]*"}, "bash", true},
		{"plus is literal", []string{"g++"}, "g++", true},
		{"plus does not act as a quantifier", []string{"g++"}, "g", false},
		{"dot is literal", []string{"a.b"}, "axb", false},
		{"re-include", []string{"ubuntu-*", "!ubuntu-desktop"}, "ubuntu-desktop", false},
		{"re-include leaves the rest ignored", []string{"ubuntu-*", "!ubuntu-desktop"}, "ubuntu-server", true},
		{"last match wins: ignored again", []string{"a*", "!ab", "ab"}, "ab", true},
		{"negation alone ignores nothing", []string{"!git"}, "git", false},
		{"negation before the ignore is overridden", []string{"!git", "g*"}, "git", true},
		{"comment", []string{"# git"}, "git", false},
		{"comment with the name after", []string{"#git"}, "git", false},
		{"escaped hash", []string{`\#odd`}, "#odd", true},
		{"escaped bang", []string{`\!odd`}, "!odd", true},
		{"blank lines", []string{"", "  ", "git"}, "git", true},
		{"surrounding spaces", []string{"  git  "}, "git", true},
		{"case sensitive", []string{"Git"}, "git", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := parseIgnore(tt.lines)
			if err != nil {
				t.Fatal(err)
			}
			if got := g.Match(tt.pkg); got != tt.want {
				t.Errorf("parseIgnore(%q).Match(%q) = %v, want %v", tt.lines, tt.pkg, got, tt.want)
			}
		})
	}
}

func TestIgnoreInvalid(t *testing.T) {
	for _, l := range []string{"[", "[a", `\`, "!["} {
		if _, err := parseIgnore([]string{l}); err == nil {
			t.Errorf("parseIgnore(%q): want error", l)
		}
	}
}

// Properties that must hold for any input; the only fixed expectation is
// that arbitrary text never panics.
func FuzzIgnore(f *testing.F) {
	f.Add("ubuntu-*", "ubuntu-desktop")
	f.Add("!a?", "ab")
	f.Add("[a-", "a")
	f.Add(`\`, "x")
	f.Fuzz(func(t *testing.T, pattern, name string) {
		g, err := parseIgnore([]string{pattern})
		if err == nil {
			g.Match(name)
		}

		// A plain name (no metacharacters) ignores itself, and a following
		// "!name" re-includes it.
		if name == "" || strings.ContainsAny(name, `*?[\`) || strings.TrimSpace(name) != name ||
			name[0] == '#' || name[0] == '!' || strings.Contains(name, "/") {
			return
		}
		g, err = parseIgnore([]string{name})
		if err != nil || !g.Match(name) {
			t.Fatalf("%q does not ignore itself (err=%v)", name, err)
		}
		g, err = parseIgnore([]string{name, "!" + name})
		if err != nil || g.Match(name) {
			t.Fatalf("%q is still ignored after !%[1]s (err=%v)", name, err)
		}
	})
}
