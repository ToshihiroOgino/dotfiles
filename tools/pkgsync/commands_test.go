package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
	oldInit := aptInitialStatus
	aptInitialStatus = filepath.Join(t.TempDir(), "missing.gz")
	t.Cleanup(func() { output, detect, aptInitialStatus = oldOut, oldDetect, oldInit })
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

func TestAptSnapshotSkipsInitial(t *testing.T) {
	useFake(t, nil, aptCmds())
	status := "Package: bash\nStatus: install ok installed\nArchitecture: amd64\n\n" +
		"Package: bat\nStatus: deinstall ok config-files\nArchitecture: amd64\n\n" +
		"Package: libc6\nStatus: install ok installed\nArchitecture: i386\n"
	var buf bytes.Buffer
	z := gzip.NewWriter(&buf)
	z.Write([]byte(status))
	z.Close()
	if err := os.WriteFile(aptInitialStatus, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := apt{}.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// bash is dropped; bat was only config-files, so it stays.
	if want := []string{"bat", "libfoo", "zsh"}; !reflect.DeepEqual(s.Packages, want) {
		t.Errorf("Packages = %q, want %q", s.Packages, want)
	}
	if want := []string{"foo-local"}; !reflect.DeepEqual(s.External, want) {
		t.Errorf("External = %q, want %q", s.External, want)
	}
	if want := []string{"bash", "libc6:i386"}; !reflect.DeepEqual(s.Initial, want) {
		t.Errorf("Initial = %q, want %q", s.Initial, want)
	}
	// A tracked package that is only present as initial is not a removal.
	d, err := compare(apt{}, Tracked{Packages: []string{"bash", "zsh"}}, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.DelPkgs) != 0 {
		t.Errorf("DelPkgs = %q, want none", d.DelPkgs)
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

func TestApplyAddsAndDropsAndClearsPending(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	writeTracked(t, root, "apt", Tracked{
		Comment:  "keep me",
		Packages: []string{"bash", "old"},
		Ignore:   []string{"lib*", "!libfoo"},
	})

	if err := cmdDiff(root, false); err != nil {
		t.Fatal(err)
	}
	if got := pending(root, "summary"); got != "+3-1\n" {
		t.Errorf("diff must refresh pending: summary = %q", got)
	}

	// libc6 is not in Packages (external); libfoo is re-included.
	if err := cmdApply(root); err != nil {
		t.Fatal(err)
	}
	tr := readTracked(t, root, "apt")
	if want := []string{"bash", "bat", "libfoo", "zsh"}; !reflect.DeepEqual(tr.Packages, want) {
		t.Errorf("Packages = %q, want %q", tr.Packages, want)
	}
	if tr.Comment != "keep me" || !reflect.DeepEqual(tr.Ignore, []string{"lib*", "!libfoo"}) {
		t.Errorf("comment/ignore lost: %+v", tr)
	}
	if tr.Buckets != nil {
		t.Errorf("apt must not get a buckets key: %v", *tr.Buckets)
	}
	// apply clears the pending files, so the prompt updates at once.
	for _, n := range []string{"apt.diff", "summary"} {
		if got := pending(root, n); got != "<none>" {
			t.Errorf("after apply: %s = %q", n, got)
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
	if err := cmdDiff(root, true); err != nil {
		t.Fatal(err)
	}
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("summary = %q, want none", got)
	}
}

func TestScoopAddWritesBucketsAndWarns(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{scoop{}}, map[string]string{"scoop export": scoopOut})

	if err := cmdApply(root); err != nil {
		t.Fatal(err)
	}
	tr := readTracked(t, root, "scoop")
	if tr.Buckets == nil || !reflect.DeepEqual(*tr.Buckets, []string{"extras", "main"}) {
		t.Errorf("Buckets = %v", tr.Buckets)
	}
	if want := []string{"git", "held"}; !reflect.DeepEqual(tr.Packages, want) {
		t.Errorf("Packages = %q, want %q", tr.Packages, want)
	}
	// apply clears pending; the next diff brings back the warning.
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("after apply: summary = %q, want none", got)
	}
	// Nothing left to add; only the failed install is a warning (a hold is not).
	if err := cmdDiff(root, false); err != nil {
		t.Fatal(err)
	}
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
	if err := cmdDiff(root, true); err != nil {
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

	if err := cmdDiff(root, true); err != nil {
		t.Fatal(err)
	}
	first := f.calls
	if first == 0 {
		t.Fatal("first check ran no command")
	}

	if err := cmdDiff(root, true); err != nil {
		t.Fatal(err)
	}
	if f.calls != first {
		t.Errorf("check within the interval ran %d more commands", f.calls-first)
	}

	if err := cmdDiff(root, false); err != nil {
		t.Fatal(err)
	}
	if f.calls != 2*first {
		t.Errorf("forced check ran %d commands, want %d", f.calls-first, first)
	}
	f.calls = first

	old := time.Now().Add(-2 * checkInterval)
	if err := os.Chtimes(filepath.Join(root, "packages", "pending", ".stamp"), old, old); err != nil {
		t.Fatal(err)
	}
	if err := cmdDiff(root, true); err != nil {
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

func TestDiffLogsFailureAndKeepsStamp(t *testing.T) {
	root := t.TempDir()
	f := useFake(t, []Manager{apt{}}, aptCmds())
	f.fail = true

	if err := cmdDiff(root, true); err == nil {
		t.Fatal("a failed run returned nil")
	}
	logs, _ := filepath.Glob(filepath.Join(root, "log", "pkgsync", "*.log"))
	if len(logs) != 1 || !regexp.MustCompile(`/\d{8}T\d{6}Z\.log$`).MatchString(logs[0]) {
		t.Fatalf("logs = %q", logs)
	}
	if b, _ := os.ReadFile(logs[0]); !strings.Contains(string(b), "boom") {
		t.Errorf("log = %q", b)
	}
	stamp := filepath.Join(root, "packages", "pending", ".stamp")
	if _, err := os.Stat(stamp); err == nil {
		t.Error("stamp written after a failure")
	}

	// Not throttled while failing, and the next plain diff reports the log.
	calls := f.calls
	if err := cmdDiff(root, true); err == nil || f.calls == calls {
		t.Errorf("a failing run was throttled (err = %v)", err)
	}
	if got := captureStderr(t, func() { cmdDiff(root, false) }); !strings.Contains(got, "boom") {
		t.Errorf("stderr = %q, want the logged error", got)
	}

	// A success updates the stamp, and the old log is no longer reported.
	f.fail = false
	if err := cmdDiff(root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stamp); err != nil {
		t.Error("stamp not written after success")
	}
	// Equal mtimes can hide the order on coarse filesystems.
	old := time.Now().Add(-time.Hour)
	all, _ := filepath.Glob(filepath.Join(root, "log", "pkgsync", "*.log"))
	for _, l := range all {
		os.Chtimes(l, old, old)
	}
	if got := captureStderr(t, func() { cmdDiff(root, false) }); strings.Contains(got, "boom") {
		t.Errorf("stale log reported: %q", got)
	}
}

func TestFailureLogNamesDoNotCollide(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 3; i++ {
		if err := writeFailureLog(root, errors.New("x")); err != nil {
			t.Fatal(err)
		}
	}
	logs, _ := filepath.Glob(filepath.Join(root, "log", "pkgsync", "*.log"))
	if len(logs) != 3 {
		t.Fatalf("logs = %q, want 3 distinct files", logs)
	}
	// Within the same second the highest sequence is the newest. The clock
	// may tick in between, so only check when all three share a second.
	if base := strings.TrimSuffix(filepath.Base(logs[0]), ".log"); strings.Count(strings.Join(logs, ""), base[:16]) == 3 {
		if got := filepath.Base(latestLog(logDir(root))); !strings.HasSuffix(got, "-2.log") {
			t.Errorf("latest = %s, want the -2 file", got)
		}
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	b, _ := io.ReadAll(r)
	return string(b)
}

func TestFailedManagerDoesNotBlockTheOther(t *testing.T) {
	root := t.TempDir()
	// scoop export is missing from the fake, so scoop fails.
	useFake(t, []Manager{scoop{}, apt{}}, aptCmds())

	err := cmdApply(root)
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

	if err := cmdApply(root); !errors.Is(err, errLocked) {
		t.Fatalf("err = %v, want errLocked", err)
	}
	if _, err := os.Stat(filepath.Join(root, "packages", "apt.json")); err == nil {
		t.Error("apt.json written while locked")
	}
	for _, throttle := range []bool{false, true} {
		if err := cmdDiff(root, throttle); !errors.Is(err, errLocked) {
			t.Fatalf("diff (throttle=%v): err = %v, want errLocked", throttle, err)
		}
	}
	if got := pending(root, "summary"); got != "<none>" {
		t.Errorf("diff wrote while locked: summary = %q", got)
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

func TestCompareAutoInstalledIsPresent(t *testing.T) {
	useFake(t, nil, aptCmds())
	s, err := apt{}.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// libfoo is installed (auto or not); tracked + installed is not a removal.
	d, err := compare(apt{}, Tracked{Packages: []string{"libfoo", "missing-pkg"}}, s)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"missing-pkg"}; !reflect.DeepEqual(d.DelPkgs, want) {
		t.Errorf("DelPkgs = %q, want %q", d.DelPkgs, want)
	}
}

func TestOldFailureLogsAreDeleted(t *testing.T) {
	root := t.TempDir()
	dir := logDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxLogs+9; i++ {
		name := fmt.Sprintf("20000101T0000%02dZ.log", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("old\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeFailureLog(root, errors.New("new")); err != nil {
		t.Fatal(err)
	}
	logs, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(logs) != maxLogs {
		t.Fatalf("%d logs, want %d", len(logs), maxLogs)
	}
	if b, _ := os.ReadFile(latestLog(dir)); string(b) != "new\n" {
		t.Errorf("newest log = %q, want the new one", b)
	}
	// The oldest are the ones that went.
	if _, err := os.Stat(filepath.Join(dir, "20000101T000000Z.log")); err == nil {
		t.Error("the oldest log was kept")
	}
	if _, err := os.Stat(filepath.Join(dir, fmt.Sprintf("20000101T0000%02dZ.log", maxLogs+8))); err != nil {
		t.Error("a recent log was deleted")
	}
}

func TestThrottledRunSkipsTheLock(t *testing.T) {
	root := t.TempDir()
	useFake(t, []Manager{apt{}}, aptCmds())
	if err := cmdDiff(root, false); err != nil {
		t.Fatal(err)
	}
	release, err := acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	// The stamp is fresh, so no lock is needed.
	if err := cmdDiff(root, true); err != nil {
		t.Errorf("throttled run failed while locked: %v", err)
	}
}
