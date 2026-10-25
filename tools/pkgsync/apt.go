package main

import (
	"runtime"
	"strings"
)

type apt struct{}

func (apt) Name() string     { return "apt" }
func (apt) HasBuckets() bool { return false }
func (apt) Available() bool  { return runtime.GOOS == "linux" && available("apt-mark") }

func (apt) Snapshot() (Snapshot, error) {
	manual, err := output("apt-mark", "showmanual")
	if err != nil {
		return Snapshot{}, err
	}
	installed, err := output("apt", "list", "--installed")
	if err != nil {
		return Snapshot{}, err
	}

	// "name/suite version arch [installed,local]" marks packages that no
	// repository provides (.deb installs). apt-mark prints foreign
	// architectures as "name:arch", so both spellings are registered.
	local := map[string]bool{}
	for _, l := range lines(installed) {
		if !strings.Contains(l, ",local]") {
			continue
		}
		name, rest, ok := strings.Cut(l, "/")
		if !ok {
			continue
		}
		local[name] = true
		if f := strings.Fields(rest); len(f) >= 3 {
			local[name+":"+f[2]] = true
		}
	}

	var s Snapshot
	for _, name := range lines(manual) {
		if local[name] {
			s.External = append(s.External, name)
		} else {
			s.Packages = append(s.Packages, name)
		}
	}
	return s, nil
}

func (apt) InstallHints(d Diff) []string {
	if len(d.DelPkgs) == 0 {
		return nil
	}
	return []string{"sudo apt install " + strings.Join(d.DelPkgs, " ")}
}
