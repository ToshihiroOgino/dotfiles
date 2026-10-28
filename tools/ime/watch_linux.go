//go:build linux

package main

import (
	"bufio"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// watch follows fcitx5's CurrentIM signal through a resident dbus-monitor, so
// nothing runs while the input method stays the same. Every sessionInterval
// without a signal the state is read once more with fcitx5-remote, which
// covers changes made while dbus-monitor was starting or down.
//
// serverPID is only used on Windows and ignored here.
func watch(socket string, _ uint32) error {
	return runWatch(socket, func() func(time.Duration) string {
		return newSignalSource().wait
	})
}

const currentIMMatch = "type='signal',interface='org.fcitx.Fcitx.InputContext1',member='CurrentIM'"

var stringLine = regexp.MustCompile(`^\s+string "(.*)"$`)

type signalSource struct {
	mu     sync.Mutex
	state  string
	notify chan struct{}
	first  bool
}

func newSignalSource() *signalSource {
	s := &signalSource{notify: make(chan struct{}, 1), first: true}
	s.resync()
	go s.run()
	return s
}

// resync reads the state with fcitx5-remote.
func (s *signalSource) resync() {
	state, err := imeState()
	if err != nil {
		state = "?"
	}
	s.set(state)
}

func (s *signalSource) set(state string) {
	s.mu.Lock()
	changed := s.state != state
	s.state = state
	s.mu.Unlock()
	if changed {
		select {
		case s.notify <- struct{}{}:
		default:
		}
	}
}

// wait is called from the watch loop only.
func (s *signalSource) wait(d time.Duration) string {
	if !s.first {
		select {
		case <-s.notify:
		case <-time.After(d):
			s.resync()
		}
	}
	s.first = false
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// run keeps a dbus-monitor alive and feeds the new input method name into the
// state. A CurrentIM signal carries (description, name, language).
func (s *signalSource) run() {
	// Pdeathsig is tied to the thread that starts dbus-monitor, so keep this
	// goroutine on one thread for good.
	runtime.LockOSThread()
	for {
		s.monitor()
		time.Sleep(sessionInterval)
	}
}

func (s *signalSource) monitor() {
	cmd := exec.Command("dbus-monitor", "--session", currentIMMatch)
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	out, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		return
	}
	defer func() {
		cmd.Process.Kill() // also when the scan stopped early
		cmd.Wait()
	}()

	sc := bufio.NewScanner(out)
	strs := -1 // strings seen in the current CurrentIM signal, -1 outside one
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "signal "):
			strs = -1
			if strings.Contains(line, "member=CurrentIM") {
				strs = 0
			}
		case strs >= 0:
			if m := stringLine.FindStringSubmatch(line); m != nil {
				if strs++; strs == 2 {
					s.set(modeOf(m[1]))
					strs = -1
				}
			}
		}
	}
}
