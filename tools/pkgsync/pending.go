package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// putFile writes content to path, or removes path when content is empty.
func putFile(path, content string) error {
	if content == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeFile(path, []byte(content))
}

// updateSummary writes <dir>/summary, a single line such as "+3-1!2"
// (added, removed, warnings) that the shell prompt prints as is. It is
// derived from the *.diff and *.warn files, so a manager that failed to
// refresh keeps its previous counts. No summary file means nothing to show.
func updateSummary(dir string) error {
	var add, del, warn int
	diffs, _ := filepath.Glob(filepath.Join(dir, "*.diff"))
	for _, p := range diffs {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, l := range lines(b) {
			switch {
			case strings.HasPrefix(l, "+ "):
				add++
			case strings.HasPrefix(l, "- "):
				del++
			}
		}
	}
	warns, _ := filepath.Glob(filepath.Join(dir, "*.warn"))
	for _, p := range warns {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		warn += len(lines(b))
	}

	var parts []string
	if add > 0 {
		parts = append(parts, fmt.Sprintf("+%d", add))
	}
	if del > 0 {
		parts = append(parts, fmt.Sprintf("-%d", del))
	}
	if warn > 0 {
		parts = append(parts, fmt.Sprintf("!%d", warn))
	}
	content := strings.Join(parts, "")
	if content != "" {
		content += "\n"
	}
	return putFile(filepath.Join(dir, "summary"), content)
}
