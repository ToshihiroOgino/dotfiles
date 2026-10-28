//go:build !windows

package tmuxctl

import "os/exec"

func hideConsole(*exec.Cmd) {}
