package main

import (
	"bufio"
	"compress/gzip"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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

// aptSourcesDir holds the apt source files. It is a variable so that tests
// can substitute it.
var aptSourcesDir = "/etc/apt/sources.list.d"

var ppaURI = regexp.MustCompile(`^https?://ppa\.(?:launchpadcontent|launchpad)\.net/([^/]+)/([^/]+)/ubuntu/?$`)

// ppas returns the enabled PPAs as "ppa:owner/name", from both the deb822
// (*.sources) and the one-line (*.list) formats. Other repositories are not
// reported.
func ppas() []string {
	set := map[string]bool{}
	add := func(uri string) {
		if m := ppaURI.FindStringSubmatch(uri); m != nil {
			set["ppa:"+m[1]+"/"+m[2]] = true
		}
	}

	sources, _ := filepath.Glob(filepath.Join(aptSourcesDir, "*.sources"))
	for _, p := range sources {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		// Stanzas are separated by blank lines; "Enabled: no" turns one off.
		for _, stanza := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n\n") {
			var uris []string
			enabled := true
			for _, l := range strings.Split(stanza, "\n") {
				k, v, ok := strings.Cut(l, ":")
				if !ok || strings.HasPrefix(l, "#") {
					continue
				}
				switch strings.ToLower(strings.TrimSpace(k)) {
				case "uris":
					uris = append(uris, strings.Fields(v)...)
				case "enabled":
					enabled = strings.TrimSpace(v) != "no"
				}
			}
			if enabled {
				for _, u := range uris {
					add(u)
				}
			}
		}
	}

	list, _ := filepath.Glob(filepath.Join(aptSourcesDir, "*.list"))
	for _, p := range list {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, l := range lines(b) {
			f := strings.Fields(l)
			if len(f) < 2 || f[0] != "deb" {
				continue
			}
			// Skip the options, as in "deb [arch=amd64] URI suite".
			i := 1
			if strings.HasPrefix(f[i], "[") {
				for i < len(f) && !strings.HasSuffix(f[i], "]") {
					i++
				}
				i++
			}
			if i < len(f) {
				add(f[i])
			}
		}
	}

	out := make([]string, 0, len(set))
	for x := range set {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

type apt struct{}

func (apt) Name() string    { return "apt" }
func (apt) Available() bool { return runtime.GOOS == "linux" && available("apt-mark") }

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
	var all []string
	for _, l := range lines(installed) {
		name, rest, ok := strings.Cut(l, "/")
		if !ok {
			continue
		}
		all = append(all, name)
		f := strings.Fields(rest)
		if len(f) >= 3 {
			all = append(all, name+":"+f[2])
		}
		if !strings.Contains(l, ",local]") {
			continue
		}
		local[name] = true
		if len(f) >= 3 {
			local[name+":"+f[2]] = true
		}
	}

	initial := initialPackages()
	s := Snapshot{Installed: all, Repos: ppas()}
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
	var cmds []string
	for _, r := range d.DelRepos {
		cmds = append(cmds, "sudo add-apt-repository -y "+r)
	}
	if len(d.DelPkgs) > 0 {
		cmds = append(cmds, "sudo apt install "+strings.Join(d.DelPkgs, " "))
	}
	return cmds
}
