//go:build darwin || linux

package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCtlRunSkipsGreetingAndNotifications(t *testing.T) {
	sr, sw := io.Pipe() // tmux -> us
	cr, cw := io.Pipe() // us -> tmux
	c := newCtl(sr, cw)
	go func() {
		io.WriteString(sw, "%begin 1 1 0\n%end 1 1 0\n%session-changed $0 a\n")
		sc := bufio.NewScanner(cr)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "bad") {
				io.WriteString(sw, "%begin 2 2 1\nparse error\n%error 2 2 1\n")
				continue
			}
			io.WriteString(sw, "%output %1 xx\n%begin 2 2 1\nclient-1\troot\n%end 2 2 1\n")
		}
		io.WriteString(sw, "%exit\n")
		sw.Close()
	}()

	out, err := c.run("list-clients")
	if err != nil || out != "client-1\troot" {
		t.Fatalf("run = %q, %v", out, err)
	}
	if _, err := c.run("bad"); err == nil || !strings.Contains(err.Error(), "parse error") {
		t.Fatalf("error = %v, want the tmux error", err)
	}
	c.close() // the server side sees EOF and ends the stream
	if _, err := c.run("list-clients"); err == nil {
		t.Error("run on a closed connection succeeded")
	}
	if c.alive() {
		<-c.done
	}
}

func TestCtlPoolWithRealTmux(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed")
	}
	t.Setenv("TMUX", "")
	const sock = "ime-go-test"
	tmux := tmuxClient{socket: sock}
	if _, err := tmux.run("-f", "/dev/null", "new-session", "-d", "-s", "t"); err != nil {
		t.Skipf("cannot start tmux: %v", err)
	}
	t.Cleanup(func() { tmux.run("kill-server") })

	pool := &ctlPool{socket: sock}
	defer pool.close()
	table := pool.table(func(string) (string, error) { return "", fmt.Errorf("fallback used") })
	out, err := pool.run(`list-clients -F "` + clientFormat + `"`)
	if err != nil || !strings.Contains(out, "\troot") {
		t.Fatalf("list-clients = %q, %v", out, err)
	}
	name, _, _ := strings.Cut(out, "\t")
	if got, err := table(name); err != nil || got != "root" {
		t.Errorf("table(%q) = %q, %v", name, got, err)
	}
	// After the server is gone the lookup falls back instead of hanging.
	tmux.run("kill-server")
	if _, err := table(name); err == nil || !strings.Contains(err.Error(), "fallback") {
		t.Errorf("after kill-server: err = %v, want the fallback", err)
	}
}

func TestCtlPoolWithTMUXSet(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed")
	}
	t.Setenv("TMUX", "")
	const sock = "ime-go-test2"
	tmux := tmuxClient{socket: sock}
	if _, err := tmux.run("-f", "/dev/null", "new-session", "-d", "-s", "t"); err != nil {
		t.Skipf("cannot start tmux: %v", err)
	}
	t.Cleanup(func() { tmux.run("kill-server") })
	info, err := tmux.run("display-message", "-p", "#{socket_path},#{pid},0")
	if err != nil {
		t.Fatal(err)
	}

	// The watcher is started by run-shell, so $TMUX names the server and
	// there is no -L.
	t.Setenv("TMUX", strings.TrimSpace(info))
	pool := &ctlPool{}
	defer pool.close()
	out, err := pool.run(`list-clients -F "` + clientFormat + `"`)
	if err != nil || !strings.Contains(out, "\troot") {
		t.Fatalf("list-clients = %q, %v", out, err)
	}
}

func TestCtlPoolBacksOffAfterFailedDial(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed")
	}
	t.Setenv("TMUX", "")
	pool := &ctlPool{socket: "ime-go-test-none"} // no server there
	if c := pool.get(); c != nil {
		t.Fatal("connected to a server that does not exist")
	}
	if pool.retry.IsZero() {
		t.Error("no backoff after a failed dial")
	}
	start := time.Now()
	for i := 0; i < 100; i++ {
		pool.get()
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("get during the backoff took %v, it must not dial", d)
	}
}
