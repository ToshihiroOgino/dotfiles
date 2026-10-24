# dotfiles commands. The recipes need sh (Git Bash on Windows).

.DEFAULT_GOAL := link

# link, link-<tool>: symlinks. Dry-run unless APPLY=1.
include scripts/link.mk

.PHONY: imestat

imestat:
	"$(MAKE)" -C tools/imestat build
