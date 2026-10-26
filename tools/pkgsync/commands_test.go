package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	aptManual = "bash\nbat\nfoo-local\nlibc6:i386\nlibfoo\nzsh\n"
	aptList   = "Listing...\n" +
		"bash/now 5.2 amd64 [installed]\n" +
		"bat/noble 0.24 amd64 [installed]\n" +
		"foo-local/now 1.0 amd64 [installed,local]\n" +
		"libc6/now 2.39 i386 [installed,local]\n" +
		"libfoo/noble 1.0 amd64 [installed]\n" +
		"no-slash-line [installed,local]\n" +
		"zsh/noble 5.9 amd64 [installed]\n"

	scoopOut = `{
  "buckets": [{"Name": "main"}, {"Name": "extras"}],
  "apps": [
    {"Name": "git",    "Source": "main",   "Info": ""},
    {"Name": "held",   "Source": "extras", "Info": "Held package"},
    {"Name": "broken",                      "Info": "Install failed"}
  ]
}`
)

// fake replaces command execution. A command that is not in out fails.
type fake struct {
	out   map[string]string
	fail  bool
	calls int
}

func useFake(t *testing.T, ms []Manager, out map[string]string) *fake {
	t.Helper()
	f := &fake{out: out}
	oldOut, oldDetect := output, detect
	output = func(name string, args ...string) ([]byte, error) {
		f.calls++
		if f.fail {
			return nil, errors.New("boom")
		}
		key := strings.Join(append([]string{name}, args...), " ")
		s, ok := f.out[key]
		if !ok {
			return nil, fmt.Errorf("unexpected command %q", key)
		}
		return []byte(s), nil
	}
	detect = func() []Manager { return ms }
	t.Cleanup(func() { output, detect = oldOut, oldDetect })
	return f
}

func aptCmds() map[string]string {
	return map[string]string{
		"apt-mark showmanual":  aptManual,
		"apt list --installed": aptList,
	}
}

func writeTracked(t *testing.T, root, mgr string, tr Tracked) {
	t.Helper()
	b, _ := json.Marshal(tr)
	p := filepath.Join(root, "packages", mgr+".json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTracked(t *testing.T, root, mgr string) Tracked {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "packages", mgr+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var tr Tracked
	if err := json.Unmarshal(b, &tr); err != nil {
		t.Fatal(err)
	}
	return tr
}

// pending returns the contents of packages/pending/<name>, or "<none>".
func pending(root, name string) string {
	b, err := os.ReadFile(filepath.Join(root, "packages", "pending", name))
	if err != nil {
		return "<none>"
	}
	return string(b)
}

func TestAptSnapshot(t *testing.T) {
	useFake(t, nil, aptCmds())
	s, err := apt{}.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"bash", "bat", "libfoo", "zsh"}; !reflect.DeepEqual(s.Packages, want) {
		t.Errorf("Packages = %q, want %q", s.Packages, want)
	}
	// libc6:i386 is foreign-architecture; apt list prints it as libc6 with
	// the architecture in the third field.
	if want := []string{"foo-local", "libc6:i386"}; !reflect.DeepEqual(s.External, want) {
		t.Errorf("External = %q, want %q", s.External, want)
	}
}

func TestScoopSnapshotCommand(t *testing.T) {
	useFake(t, nil, map[string]string{"scoop export": scoopOut})
	s, err := scoop{}.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"git", "held"}; !reflect.DeepEqual(s.Packages, want) {
		t.Errorf("Packages = %q, want %q", s.Packages, want)
	}
	if want := []string{"broken: Install failed"}; !reflect.DeepEqual(s.Warnings, want) {
		t.Errorf("Warnings = %q, want %q", s.Warnings, want)
	}
}

func TestAddThenSync(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	writeTracked(t, root, "apt", Tracked{
		Comment:  "keep me",
		Packages: []string{"bash", "old"},
		Ignore:   []string{"lib*", "!libfoo"},
	})

	if err := cmdDiff(root); err != nil {
		t.Fatal(err)
	}
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("diff must not write: summary = %q", got)
	}

	// add: new packages are written, the removal stays pending.
	// libc6 is not in Packages (external); libfoo is re-included.
	if err := cmdApply(root, false); err != nil {
		t.Fatal(err)
	}
	tr := readTracked(t, root, "apt")
	if want := []string{"bash", "bat", "libfoo", "old", "zsh"}; !reflect.DeepEqual(tr.Packages, want) {
		t.Errorf("after add: Packages = %q, want %q", tr.Packages, want)
	}
	if tr.Comment != "keep me" || !reflect.DeepEqual(tr.Ignore, []string{"lib*", "!libfoo"}) {
		t.Errorf("comment/ignore lost: %+v", tr)
	}
	if tr.Buckets != nil {
		t.Errorf("apt must not get a buckets key: %v", *tr.Buckets)
	}
	if got := pending(root, "apt.diff"); got != "- old\n" {
		t.Errorf("after add: apt.diff = %q", got)
	}
	if got := pending(root, "summary"); got != "-1\n" {
		t.Errorf("after add: summary = %q", got)
	}

	// sync: the removal is applied and nothing is pending.
	if err := cmdApply(root, true); err != nil {
		t.Fatal(err)
	}
	tr = readTracked(t, root, "apt")
	if want := []string{"bash", "bat", "libfoo", "zsh"}; !reflect.DeepEqual(tr.Packages, want) {
		t.Errorf("after sync: Packages = %q, want %q", tr.Packages, want)
	}
	for _, n := range []string{"apt.diff", "summary"} {
		if got := pending(root, n); got != "<none>" {
			t.Errorf("after sync: %s = %q", n, got)
		}
	}
}

func TestIgnoredTrackedEntryIsNotReportedRemoved(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	writeTracked(t, root, "apt", Tracked{
		Packages: []string{"bash", "bat", "libfoo", "ubuntu-minimal", "zsh"},
		Ignore:   []string{"ubuntu-*"},
	})
	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("summary = %q, want none", got)
	}
}

func TestScoopAddWritesBucketsAndWarns(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{scoop{}}, map[string]string{"scoop export": scoopOut})

	if err := cmdApply(root, true); err != nil {
		t.Fatal(err)
	}
	tr := readTracked(t, root, "scoop")
	if tr.Buckets == nil || !reflect.DeepEqual(*tr.Buckets, []string{"extras", "main"}) {
		t.Errorf("Buckets = %v", tr.Buckets)
	}
	if want := []string{"git", "held"}; !reflect.DeepEqual(tr.Packages, want) {
		t.Errorf("Packages = %q, want %q", tr.Packages, want)
	}
	// Nothing left to add; only the failed install is a warning (a hold is not).
	if got := pending(root, "summary"); got != "!1\n" {
		t.Errorf("summary = %q, want !1", got)
	}
	if got := pending(root, "scoop.warn"); got != "broken: Install failed\n" {
		t.Errorf("scoop.warn = %q", got)
	}
}

func TestSummaryCombinesManagers(t *testing.T) {
	root := t.TempDir()
	cmds := aptCmds()
	cmds["scoop export"] = scoopOut
	useFake(t, []Manager{apt{}, scoop{}}, cmds)
	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	// apt: 4 packages; scoop: 2 buckets + 2 apps, and 1 warning.
	if got := pending(root, "summary"); got != "+8!1\n" {
		t.Errorf("summary = %q, want %q", got, "+8!1\n")
	}
}

func TestCheckThrottle(t *testing.T) {
	root := t.TempDir()
	f := useFake(t, []Manager{apt{}}, aptCmds())

	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	first := f.calls
	if first == 0 {
		t.Fatal("first check ran no command")
	}

	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	if f.calls != first {
		t.Errorf("check within the interval ran %d more commands", f.calls-first)
	}

	old := time.Now().Add(-2 * checkInterval)
	if err := os.Chtimes(filepath.Join(root, ".local", "pkgsync.stamp"), old, old); err != nil {
		t.Fatal(err)
	}
	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	if f.calls != 2*first {
		t.Errorf("check after the interval ran %d commands, want %d", f.calls-first, first)
	}
	// The lock is released when check returns.
	r, err := acquire(root)
	if err != nil {
		t.Fatalf("lock still held after check: %v", err)
	}
	r()
}

func TestCheckIsSilentOnFailure(t *testing.T) {
	root := t.TempDir()
	f := useFake(t, []Manager{apt{}}, aptCmds())
	f.fail = true

	if err := cmdCheck(root); err != nil {
		t.Fatalf("check must not fail: %v", err)
	}
	errFile := filepath.Join(root, ".local", "pkgsync.err")
	b, err := os.ReadFile(errFile)
	if err != nil || !strings.Contains(string(b), "boom") {
		t.Fatalf("pkgsync.err = %q, %v", b, err)
	}

	// A later successful check clears the error.
	f.fail = false
	old := time.Now().Add(-2 * checkInterval)
	os.Chtimes(filepath.Join(root, ".local", "pkgsync.stamp"), old, old)
	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(errFile); err == nil {
		t.Error("pkgsync.err not cleared")
	}
}

func TestFailedManagerDoesNotBlockTheOther(t *testing.T) {
	root := t.TempDir()
	// scoop export is missing from the fake, so scoop fails.
	useFake(t, []Manager{scoop{}, apt{}}, aptCmds())

	err := cmdApply(root, false)
	if err == nil || !strings.Contains(err.Error(), "scoop") {
		t.Fatalf("err = %v, want a scoop error", err)
	}
	if want := []string{"bash", "bat", "libfoo", "zsh"}; !reflect.DeepEqual(readTracked(t, root, "apt").Packages, want) {
		t.Error("apt was not written after scoop failed")
	}
}

func TestApplyFailsWhileLocked(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	release, err := acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	if err := cmdApply(root, false); !errors.Is(err, errLocked) {
		t.Fatalf("err = %v, want errLocked", err)
	}
	if _, err := os.Stat(filepath.Join(root, "packages", "apt.json")); err == nil {
		t.Error("apt.json written while locked")
	}
	// check just returns.
	if err := cmdCheck(root); err != nil {
		t.Fatal(err)
	}
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("check wrote while locked: summary = %q", got)
	}
}

func TestInstallHints(t *testing.T) {
	d := Diff{DelPkgs: []string{"a", "b"}, DelBuckets: []string{"extras"}, AddPkgs: []string{"x"}}
	if got, want := (apt{}).InstallHints(d), []string{"sudo apt install a b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("apt = %q, want %q", got, want)
	}
	if got := (apt{}).InstallHints(Diff{AddPkgs: []string{"x"}}); len(got) != 0 {
		t.Errorf("apt without missing packages = %q", got)
	}
	want := []string{"scoop bucket add extras", "scoop install a b"}
	if got := (scoop{}).InstallHints(d); !reflect.DeepEqual(got, want) {
		t.Errorf("scoop = %q, want %q", got, want)
	}
}

func TestDiffFormat(t *testing.T) {
	d := Diff{AddPkgs: []string{"a"}, DelPkgs: []string{"b"}, AddBuckets: []string{"x"}}
	plain := "+ bucket:x\n+ a\n- b\n"
	if got := d.String(); got != plain {
		t.Errorf("String() = %q, want %q", got, plain)
	}
	want := "\x1b[32m+ bucket:x\x1b[0m\n\x1b[32m+ a\x1b[0m\n\x1b[31m- b\x1b[0m\n"
	if got := d.format(true); got != want {
		t.Errorf("colored = %q, want %q", got, want)
	}
}

func TestColorEnabled(t *testing.T) {
	// A file that is not a terminal never gets colors.
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if colorEnabled(f) {
		t.Error("regular file got colors")
	}
	t.Setenv("NO_COLOR", "1")
	if colorEnabled(os.Stdout) {
		t.Error("NO_COLOR ignored")
	}
}

func TestInit(t *testing.T) {
	root := t.TempDir()
	// Nothing is detected, yet every supported manager gets a file.
	useFake(t, nil, nil)

	if err := cmdInit(root); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(root, "packages", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if got, want := read("apt.json"), "{\n  \"packages\": [],\n  \"ignore\": []\n}\n"; got != want {
		t.Errorf("apt.json = %q, want %q", got, want)
	}
	if got, want := read("scoop.json"), "{\n  \"buckets\": [],\n  \"packages\": [],\n  \"ignore\": []\n}\n"; got != want {
		t.Errorf("scoop.json = %q, want %q", got, want)
	}

	// An existing file is never overwritten.
	writeTracked(t, root, "apt", Tracked{Packages: []string{"keep"}})
	if err := cmdInit(root); err != nil {
		t.Fatal(err)
	}
	if got := readTracked(t, root, "apt").Packages; !reflect.DeepEqual(got, []string{"keep"}) {
		t.Errorf("apt.json was overwritten: %q", got)
	}
	// init runs no package manager command.
}

func TestTrackedButExternalIsNotRemoved(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	// foo-local is installed from a local .deb (External); libc6:i386 too.
	writeTracked(t, root, "apt", Tracked{
		Packages: []string{"bash", "bat", "foo-local", "libfoo", "zsh"},
	})
	_, _, d, err := state(root, apt{})
	if err != nil {
		t.Fatal(err)
	}
	if !d.Empty() {
		t.Errorf("diff = %q, want empty", d.String())
	}
}

func TestLoadTrackedRejectsTrailingData(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "packages", "apt.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"packages":["a"]}{"packages":["b"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTracked(root, apt{}); err == nil {
		t.Fatal("two concatenated objects were accepted")
	}
}

func TestResolveRoot(t *testing.T) {
	t.Setenv("PKGSYNC_ROOT", filepath.Join(t.TempDir(), "missing"))
	if _, err := resolveRoot(); err == nil {
		t.Error("a missing root was accepted")
	}
	dir := t.TempDir()
	t.Setenv("PKGSYNC_ROOT", dir)
	if got, err := resolveRoot(); err != nil || got != dir {
		t.Errorf("resolveRoot = %q, %v", got, err)
	}
}

func TestDiffShowsBackgroundError(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	if err := os.MkdirAll(filepath.Join(root, ".local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".local", "pkgsync.err"), []byte("boom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// cmdDiff writes the notice to stderr; capture it.
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w
	err := cmdDiff(root)
	w.Close()
	os.Stderr = old
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(r)
	if !strings.Contains(string(b), "boom") {
		t.Errorf("stderr = %q, want the saved error", b)
	}
}

func TestCommentIsOptionalAndKept(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "packages", "apt.json")
	os.MkdirAll(filepath.Dir(p), 0o755)

	// With and without the key, the file loads and the comment never matters.
	for _, in := range []string{
		`{"packages":["a"],"ignore":[]}`,
		`{"comment":"","packages":["a"],"ignore":[]}`,
		`{"comment":"note","packages":["a"],"ignore":[]}`,
	} {
		if err := os.WriteFile(p, []byte(in), 0o644); err != nil {
			t.Fatal(err)
		}
		tr, err := loadTracked(root, apt{})
		if err != nil || !reflect.DeepEqual(tr.Packages, []string{"a"}) {
			t.Errorf("%s: %+v, %v", in, tr, err)
		}
	}

	// A comment is written back only when it is not empty.
	if err := saveTracked(root, apt{}, Tracked{}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); strings.Contains(string(b), "comment") {
		t.Errorf("empty comment was written: %s", b)
	}
	if err := saveTracked(root, apt{}, Tracked{Comment: "note"}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), `"comment": "note"`) {
		t.Errorf("comment lost: %s", b)
	}
}

func TestBrewSnapshotAndHints(t *testing.T) {
	useFake(t, nil, map[string]string{
		"brew tap": "hashicorp/tap\n",
		"brew list --formula --installed-on-request --full-name -1": "bat\nhashicorp/tap/terraform\n",
		"brew list --cask -1": "ghostty\n",
	})
	s, err := brew{}.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"bat", "hashicorp/tap/terraform", "cask:ghostty"}; !reflect.DeepEqual(s.Packages, want) {
		t.Errorf("packages = %v, want %v", s.Packages, want)
	}
	if want := []string{"hashicorp/tap"}; !reflect.DeepEqual(s.Buckets, want) {
		t.Errorf("buckets = %v, want %v", s.Buckets, want)
	}
	got := brew{}.InstallHints(Diff{DelPkgs: []string{"bat", "cask:ghostty"}, DelBuckets: []string{"x/y"}})
	want := []string{"brew tap x/y", "brew install bat", "brew install --cask ghostty"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("hints = %v, want %v", got, want)
	}
}

func TestIgnoreAppliesToBuckets(t *testing.T) {
	for _, m := range []Manager{scoop{}, brew{}} {
		d, err := compare(m,
			Tracked{Buckets: &[]string{"gone"}, Ignore: []string{"extras", "gone"}},
			Snapshot{Buckets: []string{"main", "extras"}})
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{"main"}; !reflect.DeepEqual(d.AddBuckets, want) || len(d.DelBuckets) != 0 {
			t.Errorf("%s: add=%v del=%v", m.Name(), d.AddBuckets, d.DelBuckets)
		}
	}
}
