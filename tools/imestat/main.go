// Command imestat reports the IME input mode of the foreground window for a
// tmux status line.
//
//	imestat                  print the mode once: "あ", "A", or "?"
//	imestat --watch [-L s]   stay resident and push the mode into the tmux
//	                         user option @ime of every session, only when it
//	                         changes (see tools/imestat/README.md)
//
// On unsupported platforms nothing is printed to stdout; a message goes to
// stderr and the exit code is 1.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	watchMode := flag.Bool("watch", false, "stay resident and push changes into tmux @ime")
	socket := flag.String("L", "", "tmux socket name (default socket if empty)")
	serverPID := flag.Uint("server-pid", 0, "internal: psmux server to follow (set by --watch when it detaches)")
	flag.Parse()

	if *watchMode {
		if err := watch(*socket, uint32(*serverPID)); err != nil {
			fmt.Fprintln(os.Stderr, "imestat:", err)
			os.Exit(1)
		}
		return
	}

	s, err := imeState()
	if err != nil {
		fmt.Fprintln(os.Stderr, "imestat:", err)
		os.Exit(1)
	}
	fmt.Print(s)
}
