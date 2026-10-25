package main

import "os"

const (
	bold   = "1"
	red    = "31"
	green  = "32"
	yellow = "33"
	cyan   = "36"
)

// painter wraps text in ANSI colors; the zero value leaves text as is.
type painter bool

func (p painter) paint(code, s string) string {
	if !p {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// colorEnabled reports whether f is a terminal that should get colors.
// NO_COLOR (https://no-color.org) and TERM=dumb turn them off.
func colorEnabled(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
