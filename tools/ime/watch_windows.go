//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/imectl"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/tmuxctl"
)

const (
	pollInterval    = 100 * time.Millisecond // IME state check (in-process, cheap)
	sessionInterval = 2 * time.Second        // look for newly created sessions

	synchronize    = 0x00100000 // SYNCHRONIZE
	infinite       = 0xFFFFFFFF // INFINITE
	waitObject0    = 0x00000000 // WAIT_OBJECT_0
	waitAbandoned0 = 0x00000080 // WAIT_ABANDONED_0
	createNoWindow = 0x08000000 // CREATE_NO_WINDOW
	detached       = 0x00000008 // DETACHED_PROCESS
	newProcGroup   = 0x00000200 // CREATE_NEW_PROCESS_GROUP
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutex  = kernel32.NewProc("CreateMutexW")
	waitMultiple = kernel32.NewProc("WaitForMultipleObjects")
)

// serverNames are the executables psmux runs its server as.
var serverNames = map[string]bool{"tmux.exe": true, "psmux.exe": true, "pmux.exe": true}

// watch runs one resident watcher per socket. Every psmux server starts one
// (run-shell -b in tmux.conf); a named mutex lets exactly one of them work
// while the others wait to take over. Each exits when its own server exits,
// because psmux does not kill run-shell -b children on kill-server.
//
// psmux starts run-shell commands through `pwsh -Command`, and that pwsh
// (~75 MB) would live as long as we do. So the first invocation locates its
// server, re-launches itself detached with --server-pid, and exits at once.
func watch(socket string, serverPID uint32) error {
	if serverPID == 0 {
		// Without a server to follow nothing would ever stop us, e.g. when a
		// server dies while its config is still running. Refuse instead.
		pid, ok := serverProcess()
		if !ok {
			return fmt.Errorf("no psmux server among ancestors (pass --server-pid to run by hand)")
		}
		return detach(socket, pid)
	}

	// Mutex ownership is per OS thread.
	runtime.LockOSThread()

	// psmux re-runs the config when a warm server is claimed, so a server can
	// start a second watcher. Keep one per server.
	_, existed, err := namedMutex(fmt.Sprintf(`Local\ime-server-%d`, serverPID))
	if err != nil {
		return err
	}
	if existed {
		return nil
	}

	name := `Local\ime-watch`
	if socket != "" {
		name += "-" + socket
	}
	mu, _, err := namedMutex(name)
	if err != nil {
		return err
	}

	server, err := syscall.OpenProcess(synchronize, false, serverPID)
	if err != nil {
		return nil // server already gone
	}
	handles := []syscall.Handle{mu, server}

	// Block until we own the mutex, or until our server exits first.
	r, _, err := waitMultiple.Call(uintptr(len(handles)),
		uintptr(unsafe.Pointer(&handles[0])), 0, infinite)
	switch r {
	case waitObject0, waitAbandoned0:
		// Owner. An abandoned mutex means the previous owner's server exited.
	case waitObject0 + 1:
		return nil
	default:
		return fmt.Errorf("WaitForMultipleObjects: %w", err)
	}

	tmux := tmuxctl.Client{Socket: socket}
	var last string
	known := map[string]bool{}
	nextScan := time.Now().Add(sessionInterval)
	for {
		if ev, _ := syscall.WaitForSingleObject(server, 0); ev == waitObject0 {
			return nil
		}
		state, _ := imectl.State()
		now := time.Now()
		if state != last {
			// Notify tmux only when the mode changes.
			last = state
			known = push(tmux, state, known, true)
			nextScan = now.Add(sessionInterval)
		} else if now.After(nextScan) {
			// New sessions have no @ime yet; send the current mode to them only.
			known = push(tmux, state, known, false)
			nextScan = now.Add(sessionInterval)
		}
		time.Sleep(pollInterval)
	}
}

// namedMutex opens or creates a named mutex without owning it. existed reports
// whether another process had already created it. The handle stays open for
// the life of the process.
func namedMutex(name string) (h syscall.Handle, existed bool, err error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0, false, err
	}
	r, _, callErr := createMutex.Call(0, 0, uintptr(unsafe.Pointer(namePtr)))
	if r == 0 {
		return 0, false, fmt.Errorf("CreateMutex %s: %w", name, callErr)
	}
	return syscall.Handle(r), callErr == syscall.ERROR_ALREADY_EXISTS, nil
}

// detach re-launches this executable as a detached watcher that follows pid.
func detach(socket string, pid uint32) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{"watch", "--server-pid", fmt.Sprint(pid)}
	if socket != "" {
		args = append(args, "-L", socket)
	}
	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detached | newProcGroup}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// serverProcess finds the nearest ancestor that is a psmux server. run-shell
// children get an empty TMUX and no format expansion, so the process tree is
// the only reliable link back to the server.
func serverProcess() (uint32, bool) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, false
	}
	defer syscall.CloseHandle(snap)

	type proc struct {
		parent uint32
		exe    string
	}
	procs := map[uint32]proc{}
	var e syscall.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = syscall.Process32First(snap, &e); err == nil; err = syscall.Process32Next(snap, &e) {
		procs[e.ProcessID] = proc{e.ParentProcessID, strings.ToLower(syscall.UTF16ToString(e.ExeFile[:]))}
	}

	pid := uint32(os.Getppid())
	for depth := 0; depth < 16 && pid != 0; depth++ {
		p, ok := procs[pid]
		if !ok {
			return 0, false
		}
		if serverNames[p.exe] {
			return pid, true
		}
		pid = p.parent
	}
	return 0, false
}
