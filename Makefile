# dotfiles commands. The recipes need sh (Git Bash on Windows).

.DEFAULT_GOAL := link

include scripts/link.mk

.PHONY: imestat

imestat:
	"$(MAKE)" -C tools/imestat build
