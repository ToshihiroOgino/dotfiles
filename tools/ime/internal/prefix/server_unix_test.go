//go:build darwin || linux

package prefix

import (
	"os"
	"testing"
)

// The tests never send a request: handling one would switch the real IME.

func TestNotifyWithoutServer(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	if Notify("none", "client-1") {
		t.Error("Notify succeeded with no server")
	}
}

func TestServeCreatesAndRemovesSocket(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)
	stop := Serve("t")
	path, err := socketPath("t")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("socket not created: %v", err)
	}
	stop()
	if _, err := os.Stat(path); err == nil {
		t.Error("socket left behind after stop")
	}
}

func TestSocketDirIsPrivateWithoutXDG(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("TMPDIR", t.TempDir())
	d, err := socketDir()
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(d)
	if err != nil || fi.Mode().Perm()&0o077 != 0 {
		t.Errorf("%s: %v, mode %v", d, err, fi.Mode())
	}
}
