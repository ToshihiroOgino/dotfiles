// Command ime reads and switches the IME input mode of the foreground
// window.
//
//	ime [get]                print the mode once: "あ", "A", or "?"
//	ime set en|ja            switch to alphanumeric / native (Japanese) input
//	ime prefix [-L s] CLIENT switch to alphanumeric input until the tmux client
//	                            leaves the prefix key table, then restore the mode
//	ime watch [-L s]         stay resident and push the mode into the tmux
//	                            user option @ime of every session, only when it
//	                            changes (see tools/ime/README.md)
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
	cmd, args := "get", os.Args[1:]
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	var err error
	switch cmd {
	case "get":
		var s string
		if s, err = imeState(); err == nil {
			fmt.Print(s)
		}
	case "set":
		if len(args) != 1 || (args[0] != "en" && args[0] != "ja") {
			err = fmt.Errorf("usage: ime set en|ja")
		} else {
			err = setIME(args[0] == "ja")
		}
	case "prefix":
		fs := flag.NewFlagSet("prefix", flag.ExitOnError)
		socket := fs.String("L", "", "tmux socket name (default socket if empty)")
		fs.Parse(args)
		if fs.NArg() != 1 {
			err = fmt.Errorf("usage: ime prefix [-L socket] client")
		} else {
			prefixSession(*socket, fs.Arg(0))
		}
	case "watch":
		fs := flag.NewFlagSet("watch", flag.ExitOnError)
		socket := fs.String("L", "", "tmux socket name (default socket if empty)")
		serverPID := fs.Uint("server-pid", 0, "internal: psmux server to follow (set by watch when it detaches)")
		fs.Parse(args)
		err = watch(*socket, uint32(*serverPID))
	default:
		err = fmt.Errorf("unknown command %q (get, set, prefix, watch)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ime:", err)
		os.Exit(1)
	}
}
