package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAcquireExclusive(t *testing.T) {
	root := t.TempDir()
	const n = 16
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		got     int
		release func()
	)
	start := make(chan struct{})
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			r, err := acquire(root)
			if err == nil {
				mu.Lock()
				got++
				release = r
				mu.Unlock()
			} else if !errors.Is(err, errLocked) {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got != 1 {
		t.Fatalf("%d callers got the lock, want 1", got)
	}

	release()
	r, err := acquire(root)
	if err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	r()
}

// A lock held by another process blocks acquire, and disappears with that
// process even when it is killed.
func TestAcquireAcrossProcesses(t *testing.T) {
	root := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=TestLockHolderProcess")
	cmd.Env = append(os.Environ(), "PKGSYNC_LOCK_HOLDER="+root)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()

	line, _ := bufio.NewReader(out).ReadString('\n')
	if line != "locked\n" {
		t.Fatalf("holder said %q", line)
	}
	if _, err := acquire(root); !errors.Is(err, errLocked) {
		t.Fatalf("while held: got %v, want errLocked", err)
	}

	cmd.Process.Kill()
	cmd.Wait()
	r, err := acquire(root)
	if err != nil {
		t.Fatalf("after the holder was killed: %v", err)
	}
	r()
}

// TestLockHolderProcess is the child of TestAcquireAcrossProcesses; on its
// own it does nothing.
func TestLockHolderProcess(t *testing.T) {
	root := os.Getenv("PKGSYNC_LOCK_HOLDER")
	if root == "" {
		t.Skip("helper process")
	}
	if _, err := acquire(root); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("locked")
	time.Sleep(time.Minute)
}

func TestWriteFileReplaces(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	for _, want := range []string{"one", "two"} {
		if err := writeFile(p, []byte(want)); err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(p); string(b) != want {
			t.Fatalf("got %q, want %q", b, want)
		}
	}
	if m, _ := filepath.Glob(filepath.Join(filepath.Dir(p), "*.tmp")); len(m) != 0 {
		t.Fatalf("temporary files left: %v", m)
	}
}
