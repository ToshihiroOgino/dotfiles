//go:build darwin || linux

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// watchName names the lock and the socket of the watcher for a tmux socket.
func watchName(socket string) string {
	if socket == "" {
		return "ime-watch"
	}
	return "ime-watch-" + socket
}

// socketDir returns a directory only this user can use: $XDG_RUNTIME_DIR, or
// a private one under the temporary directory. The socket must not sit in a
// shared directory, where another user could create it first or connect.
func socketDir() (string, error) {
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		return d, nil
	}
	d := filepath.Join(os.TempDir(), fmt.Sprintf("ime-%d", os.Getuid()))
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	fi, err := os.Lstat(d)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() || fi.Mode().Perm()&0o077 != 0 || !ownedByMe(fi) {
		return "", fmt.Errorf("%s is not a private directory", d)
	}
	return d, nil
}

func ownedByMe(fi os.FileInfo) bool {
	st, ok := fi.Sys().(*syscall.Stat_t)
	return ok && int(st.Uid) == os.Getuid()
}

func prefixSocketPath(socket string) (string, error) {
	d, err := socketDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, watchName(socket)+".sock"), nil
}

// servePrefix accepts `ime prefix` requests on a Unix socket and runs the
// prefix sessions in this process, over one tmux control connection. It must
// be called with the watcher lock held, which makes removing a stale socket
// safe. The returned func stops it: running sessions end first, so that they
// restore the IME.
func servePrefix(socket string) (stop func()) {
	path, err := prefixSocketPath(socket)
	if err != nil {
		return func() {} // `ime prefix` then runs the session by itself
	}
	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		return func() {}
	}
	pool := &ctlPool{socket: socket}
	lookup := pool.table(spawnTable(tmuxClient{socket: socket}))
	var closing atomic.Bool
	table := func(client string) (string, error) {
		if closing.Load() {
			return "", errors.New("watcher is stopping") // ends the session
		}
		return lookup(client)
	}
	var sessions sync.WaitGroup
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			sessions.Add(1)
			go func() {
				defer sessions.Done()
				defer func() { recover() }() // one bad session must not kill the watcher
				conn.SetReadDeadline(time.Now().Add(time.Second))
				line, err := bufio.NewReader(conn).ReadString('\n')
				conn.Close()
				if client := strings.TrimSpace(line); err == nil && client != "" {
					prefixSession(client, table)
				}
			}()
		}
	}()
	return func() {
		l.Close() // also removes the socket file
		closing.Store(true)
		done := make(chan struct{})
		go func() { sessions.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
		pool.close()
	}
}

// notifyDaemon hands the prefix session of client to the running watcher. It
// reports false when there is none, and the caller does the work itself.
func notifyDaemon(socket, client string) bool {
	path, err := prefixSocketPath(socket)
	if err != nil {
		return false
	}
	conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err = io.WriteString(conn, client+"\n")
	return err == nil
}
