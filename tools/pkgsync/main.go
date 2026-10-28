// Command pkgsync compares the manually installed packages of this machine
// with the tracked lists in packages/<mgr>.json.
//
//	pkgsync init     create an empty packages/<mgr>.json for every supported manager that has none
//	pkgsync [diff]   print the difference and refresh packages/pending/
//	                 (--throttle: skip if the last successful run is under 5 minutes old)
//	pkgsync apply    add new and drop removed packages in packages/<mgr>.json
//	pkgsync help     show the usage
//
// Package managers are detected from PATH. pkgsync never touches git and
// never installs or removes anything.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const helpText = `pkgsync compares installed packages with packages/<mgr>.json.

usage: pkgsync [command]

commands:
  init   create an empty packages/<mgr>.json for every supported manager that has none
  diff   show the difference, and install commands for what is missing here (default);
         also refreshes packages/pending/
         diff --throttle does nothing if the last successful run is under 5 minutes old
         (for shell startup; redirect the output yourself)
  apply  add new packages to the list
  help   show this help

Failures are logged to log/pkgsync/<UTC time>.log.

The dotfiles root is $PKGSYNC_ROOT, or ~/dotfiles if unset.

Nothing is installed or removed, and git is never touched.
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, helpText) }
	flag.Parse()

	cmd := "diff"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}
	rest := flag.Args()
	if len(rest) > 0 {
		rest = rest[1:]
	}
	// The only option each command takes, if any.
	opt := map[string]string{"diff": "--throttle"}[cmd]
	var optOn bool
	if opt != "" && len(rest) == 1 && rest[0] == opt {
		optOn, rest = true, nil
	}
	if len(rest) > 0 {
		fmt.Fprintf(os.Stderr, "pkgsync: unexpected arguments %q\n", rest)
		os.Exit(2)
	}

	if cmd == "help" {
		fmt.Print(helpText)
		return
	}

	root, err := resolveRoot()
	if err == nil {
		switch cmd {
		case "init":
			err = cmdInit(root)
		case "diff":
			err = cmdDiff(root, optOn)
		case "apply":
			err = cmdApply(root)
		default:
			fmt.Fprintf(os.Stderr, "pkgsync: unknown command %q\n", cmd)
			flag.Usage()
			os.Exit(2)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "pkgsync:", err)
		os.Exit(1)
	}
}

// resolveRoot returns the dotfiles root, which must exist: a typo in
// PKGSYNC_ROOT must not make init or apply create directories elsewhere.
func resolveRoot() (string, error) {
	root := os.Getenv("PKGSYNC_ROOT")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, "dotfiles")
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("dotfiles root %s not found (set PKGSYNC_ROOT)", root)
	}
	return root, nil
}
