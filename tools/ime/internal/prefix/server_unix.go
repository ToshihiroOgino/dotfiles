//go:build darwin || linux

package prefix

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

	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/tmuxctl"
)

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

func socketPath(socket string) (string, error) {
	d, err := socketDir()
	if err != nil {
		return "", err
	}
	name := "ime-prefix"
	if socket != "" {
		name += "-" + socket
	}
	return filepath.Join(d, name+".sock"), nil
}

// maxSocketPath is below the shortest sun_path (104 bytes on macOS, 108 on
// Linux), counting the terminating NUL.
const maxSocketPath = 104

// Serve accepts `ime prefix` requests on a Unix socket and runs the
// prefix sessions in this process, over one tmux control connection. It must
// be called with the watcher lock held, which makes removing a stale socket
// safe. The returned func stops it: running sessions end first, so that they
// restore the IME.
func Serve(socket string) (stop func()) {
	// Without a socket, `ime prefix` runs the session by itself.
	path, err := socketPath(socket)
	if err == nil && len(path) >= maxSocketPath {
		err = fmt.Errorf("%s is too long for a Unix socket", path)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ime: prefix sessions run per key press:", err)
		return func() {}
	}
	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ime: prefix sessions run per key press:", err)
		return func() {}
	}
	pool := &tmuxctl.Pool{Socket: socket}
	lookup := pool.Table(tmuxctl.SpawnTable(tmuxctl.Client{Socket: socket}))
	var closing atomic.Bool
	table := func(client string) (string, error) {
		if closing.Load() {
			return "", errors.New("watcher is stopping") // ends the session
		}
		return lookup(client)
	}
	var sessions sync.WaitGroup
	accepted := make(chan struct{}) // closed when no more sessions can start
	go func() {
		defer close(accepted)
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
					Session(client, table)
				}
			}()
		}
	}()
	return func() {
		l.Close()  // also removes the socket file
		<-accepted // sessions.Add must not run while Wait does
		closing.Store(true)
		done := make(chan struct{})
		go func() { sessions.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
		pool.Close()
	}
}

// Notify hands the prefix session of client to the running watcher. It
// reports false when there is none, and the caller does the work itself.
func Notify(socket, client string) bool {
	path, err := socketPath(socket)
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
