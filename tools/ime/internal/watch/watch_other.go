//go:build !windows && !darwin && !linux

package watch

import (
	"fmt"
	"runtime"
)

func Run(string, uint32) error {
	return fmt.Errorf("watch not supported on %s", runtime.GOOS)
}
