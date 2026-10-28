package tmuxctl

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ctlTimeout     = 2 * time.Second
	redialInterval = 2 * time.Second // between attempts to connect to tmux
)

// tmuxError is an error that tmux reported for a command, as opposed to a
// broken connection.
type tmuxError string

func (e tmuxError) Error() string { return "tmux: " + string(e) }

type ctlResult struct {
	out string
	err error
}

// ctlConn is one tmux control-mode connection (`tmux -C`) used as a command
// channel, so that polling does not start a tmux process per call.
type ctlConn struct {
	cmd  *exec.Cmd // nil in tests
	in   io.WriteCloser
	resp chan ctlResult
	done chan struct{} // closed when the connection ends
	shut atomic.Bool   // close was called
	mu   sync.Mutex    // one command at a time
}

func newCtl(r io.Reader, w io.WriteCloser) *ctlConn {
	c := &ctlConn{in: w, resp: make(chan ctlResult, 1), done: make(chan struct{})}
	go c.read(r)
	return c
}

// read splits the control-mode output into command responses. Notifications
// such as %output are dropped. Only a block whose flags field is 1 answers a
// command of ours; the greeting block that tmux sends on attach has 0.
func (c *ctlConn) read(r io.Reader) {
	defer close(c.done)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	var buf []string
	var inBlock, mine bool
	for sc.Scan() {
		l := sc.Text()
		switch {
		case !inBlock && strings.HasPrefix(l, "%begin "):
			f := strings.Fields(l)
			inBlock, mine, buf = true, len(f) == 4 && f[3] == "1", nil
		case inBlock && (strings.HasPrefix(l, "%end ") || strings.HasPrefix(l, "%error ")):
			inBlock = false
			if !mine {
				continue
			}
			res := ctlResult{out: strings.Join(buf, "\n")}
			if strings.HasPrefix(l, "%error ") {
				res.err = tmuxError(res.out)
			}
			select {
			case c.resp <- res:
			default:
			}
		case inBlock:
			buf = append(buf, l)
		case strings.HasPrefix(l, "%exit"):
			return
		}
	}
}

// run sends one tmux command and returns its output.
func (c *ctlConn) run(command string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.resp: // an answer that arrived after a timeout
	default:
	}
	if _, err := io.WriteString(c.in, command+"\n"); err != nil {
		return "", err
	}
	select {
	case r := <-c.resp:
		return r.out, r.err
	case <-c.done:
		return "", errors.New("tmux control connection closed")
	case <-time.After(ctlTimeout):
		c.close()
		return "", errors.New("tmux control connection timed out")
	}
}

func (c *ctlConn) alive() bool {
	if c.shut.Load() {
		return false
	}
	select {
	case <-c.done:
		return false
	default:
		return true
	}
}

func (c *ctlConn) close() {
	c.shut.Store(true)
	c.in.Close()
	if c.cmd != nil {
		c.cmd.Process.Kill()
	}
}

// dialControl starts `tmux -C attach` against socket. $TMUX, set when tmux
// started us, makes a bare `attach` find the right server.
func dialControl(socket string) (*ctlConn, error) {
	args := []string{"-C", "attach"}
	if socket != "" {
		args = append([]string{"-L", socket}, args...)
	}
	cmd := exec.Command("tmux", args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := newCtl(out, in)
	c.cmd = cmd
	go func() {
		<-c.done           // all reads are done; now Wait may close the pipe
		cmd.Process.Kill() // reading can stop with tmux still running
		cmd.Wait()
	}()
	// The control client has no size and must not count for window-size, and
	// the panes' output is of no use here. A tmux error only means an older
	// tmux; a broken connection (tmux exited right away) is a failed dial.
	if _, err := c.run("refresh-client -f no-output,ignore-size"); err != nil {
		var te tmuxError
		if !errors.As(err, &te) {
			c.close()
			return nil, err
		}
	}
	return c, nil
}

// Pool keeps one control connection, reconnecting lazily. Without one,
// callers fall back to starting tmux.
type Pool struct {
	Socket  string // tmux socket name for -L; the default socket if empty
	mu      sync.Mutex
	c       *ctlConn
	dialing bool
	retry   time.Time
}

// get returns the live connection, or nil. It never waits for a dial that
// another caller is making: that caller's poll is slow enough already.
func (p *Pool) get() *ctlConn {
	p.mu.Lock()
	if p.c != nil && p.c.alive() {
		defer p.mu.Unlock()
		return p.c
	}
	if p.c != nil {
		p.c = nil
		p.retry = time.Now().Add(redialInterval) // it died: do not redial at once
	}
	if p.dialing || time.Now().Before(p.retry) {
		p.mu.Unlock()
		return nil
	}
	p.dialing = true
	p.mu.Unlock()

	c, err := dialControl(p.Socket)

	p.mu.Lock()
	defer p.mu.Unlock()
	p.dialing = false
	if err != nil {
		p.retry = time.Now().Add(redialInterval) // e.g. no session yet
		return nil
	}
	p.c = c
	return c
}

// Run sends one tmux command over the connection and returns its output.
func (p *Pool) Run(command string) (string, error) {
	c := p.get()
	if c == nil {
		return "", errors.New("no tmux control connection")
	}
	out, err := c.run(command)
	var te tmuxError
	if err != nil && !errors.As(err, &te) {
		// A broken connection (a timeout closes it): drop it.
		c.close()
		p.mu.Lock()
		if p.c == c {
			p.c = nil
			p.retry = time.Now().Add(redialInterval)
		}
		p.mu.Unlock()
	}
	return out, err
}

// Table looks the key table up over the control connection, and with
// fallback when there is none.
func (p *Pool) Table(fallback TableFunc) TableFunc {
	return func(client string) (string, error) {
		out, err := p.Run(`list-clients -F "` + ClientFormat + `"`)
		if err != nil {
			return fallback(client)
		}
		return TableOf(out, client)
	}
}

// Close ends the connection.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.c != nil {
		p.c.close()
		p.c = nil
	}
}
