package main

import (
	"bufio"
	"compress/gzip"
	"os"
	"runtime"
	"strings"
)

// aptInitialStatus is the dpkg status the installer recorded right after
// installing the OS. It is a variable so that tests can substitute it.
var aptInitialStatus = "/var/log/installer/initial-status.gz"

// initialPackages returns the packages installed by the OS installer, as
// "name" and "name:arch". It returns nil when the file is missing or unreadable.
func initialPackages() map[string]bool {
	f, err := os.Open(aptInitialStatus)
	if err != nil {
		return nil
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil
	}
	defer z.Close()

	set := map[string]bool{}
	var name, arch string
	installed := false
	flush := func() {
		if installed && name != "" {
			set[name] = true
			if arch != "" {
				set[name+":"+arch] = true
			}
		}
		name, arch, installed = "", "", false
	}
	sc := bufio.NewScanner(z)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		l := sc.Text()
		switch {
		case l == "":
			flush()
		case strings.HasPrefix(l, "Package: "):
			name = l[len("Package: "):]
		case strings.HasPrefix(l, "Architecture: "):
			arch = l[len("Architecture: "):]
		case strings.HasPrefix(l, "Status: "):
			installed = strings.HasSuffix(l, " installed")
		}
	}
	flush()
	return set
}

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

	initial := initialPackages()
	var s Snapshot
	for _, name := range lines(manual) {
		if initial[name] {
			s.Initial = append(s.Initial, name)
			continue
		}
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
