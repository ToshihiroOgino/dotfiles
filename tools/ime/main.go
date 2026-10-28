// Command ime reads and switches the IME input mode of the foreground
// window.
//
//	ime [get] [-s]           print the mode once: "ja", "en", or "?"
//	                            (-s: the status bar symbols "あ", "A", or "?")
//	ime set en|ja            switch to alphanumeric / native (Japanese) input
//	ime prefix [-L s] CLIENT switch to alphanumeric input until the tmux client
//	                            leaves the prefix key table, then restore the mode
//	ime watch [-L s]         stay resident and push the mode into the tmux
//	                            user option @ime of every session, only when it
//	                            changes; on Linux and macOS it also runs the
//	                            prefix sessions that `ime prefix` hands over
//	                            (see tools/ime/README.md)
//
// On unsupported platforms nothing is printed to stdout; a message goes to
// stderr and the exit code is 1.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/imectl"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/prefix"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/tmuxctl"
	"github.com/ToshihiroOgino/dotfiles/tools/ime/internal/watch"
)

func main() {
	cmd, args := "get", os.Args[1:]
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	var err error
	switch cmd {
	case "get":
		fs := flag.NewFlagSet("get", flag.ExitOnError)
		symbol := fs.Bool("s", false, `print "あ"/"A" instead of "ja"/"en"`)
		fs.Parse(args)
		var s string
		if s, err = imectl.State(); err == nil {
			if !*symbol {
				switch s {
				case "あ":
					s = "ja"
				case "A":
					s = "en"
				}
			}
			fmt.Print(s)
		}
	case "set":
		if len(args) != 1 || (args[0] != "en" && args[0] != "ja") {
			err = fmt.Errorf("usage: ime set en|ja")
		} else {
			err = imectl.Set(args[0] == "ja")
		}
	case "prefix":
		fs := flag.NewFlagSet("prefix", flag.ExitOnError)
		socket := fs.String("L", "", "tmux socket name (default socket if empty)")
		fs.Parse(args)
		if fs.NArg() != 1 {
			err = fmt.Errorf("usage: ime prefix [-L socket] client")
		} else {
			// A running watcher does the work if there is one; it keeps a
			// single tmux connection instead of starting tmux per poll.
			if !prefix.Notify(*socket, fs.Arg(0)) {
				prefix.Session(fs.Arg(0), tmuxctl.SpawnTable(tmuxctl.Client{Socket: *socket}))
			}
		}
	case "watch":
		fs := flag.NewFlagSet("watch", flag.ExitOnError)
		socket := fs.String("L", "", "tmux socket name (default socket if empty)")
		serverPID := fs.Uint("server-pid", 0, "internal: psmux server to follow (set by watch when it detaches)")
		fs.Parse(args)
		err = watch.Run(*socket, uint32(*serverPID))
	default:
		err = fmt.Errorf("unknown command %q (get, set, prefix, watch)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ime:", err)
		os.Exit(1)
	}
}
