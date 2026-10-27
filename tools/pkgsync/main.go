// Command pkgsync compares the manually installed packages of this machine
// with the tracked lists in packages/<mgr>.json.
//
//	pkgsync init     create an empty packages/<mgr>.json for every supported manager that has none
//	pkgsync [diff]   print the difference; nothing is written
//	pkgsync add      add new packages to packages/<mgr>.json
//	pkgsync sync     add new and drop removed packages
//	pkgsync check    for shell startup: at most once per 5 minutes, refresh
//	                 packages/pending/<mgr>.diff (-f ignores the interval)
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
  diff   show the difference, and install commands for what is missing here (default)
  add    add new packages to the list
  sync   add new packages and drop removed ones
  check  for shell startup: refresh packages/pending/ (once per 5 minutes, silent)
         check -f refreshes now, ignoring the interval
  help   show this help

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
	force := false
	rest := flag.Args()
	if len(rest) > 0 {
		rest = rest[1:]
	}
	if cmd == "check" && len(rest) == 1 && rest[0] == "-f" {
		force, rest = true, nil
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
			err = cmdDiff(root)
		case "add":
			err = cmdApply(root, false)
		case "sync":
			err = cmdApply(root, true)
		case "check":
			err = cmdCheck(root, force)
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
// PKGSYNC_ROOT must not make init or add create directories elsewhere.
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
