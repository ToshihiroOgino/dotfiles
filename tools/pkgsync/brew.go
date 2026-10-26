package main

import (
	"strings"
)

// caskPrefix marks casks in packages/brew.json; formulae are bare names.
const caskPrefix = "cask:"

type brew struct{}

func (brew) Name() string     { return "brew" }
func (brew) HasBuckets() bool { return true }
func (brew) Available() bool  { return available("brew") }

func (brew) Snapshot() (Snapshot, error) {
	taps, err := output("brew", "tap")
	if err != nil {
		return Snapshot{}, err
	}
	leaves, err := output("brew", "leaves")
	if err != nil {
		return Snapshot{}, err
	}
	casks, err := output("brew", "list", "--cask", "-1")
	if err != nil {
		return Snapshot{}, err
	}

	s := Snapshot{Buckets: lines(taps), Packages: lines(leaves)}
	for _, c := range lines(casks) {
		s.Packages = append(s.Packages, caskPrefix+c)
	}
	return s, nil
}

func (brew) InstallHints(d Diff) []string {
	var cmds []string
	for _, t := range d.DelBuckets {
		cmds = append(cmds, "brew tap "+t)
	}
	var formulae, casks []string
	for _, p := range d.DelPkgs {
		if c, ok := strings.CutPrefix(p, caskPrefix); ok {
			casks = append(casks, c)
		} else {
			formulae = append(formulae, p)
		}
	}
	if len(formulae) > 0 {
		cmds = append(cmds, "brew install "+strings.Join(formulae, " "))
	}
	if len(casks) > 0 {
		cmds = append(cmds, "brew install --cask "+strings.Join(casks, " "))
	}
	return cmds
}
