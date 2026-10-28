.DEFAULT_GOAL := link

include scripts/env.mk
include scripts/link.mk

all: tools link external systemd-enable

.PHONY: all ime pkgsync systemd-enable external nvim tree-sitter omz claude

tools: ime pkgsync

ime:
	"$(MAKE)" -C tools/ime build

pkgsync:
	"$(MAKE)" -C tools/pkgsync build

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh

external: nvim tree-sitter omz claude

nvim:
	@sh external/nvim.sh

tree-sitter:
	@sh external/tree-sitter.sh

claude:
	@sh external/claude.sh

omz:
	@if [ "$(PLATFORM)" != windows ]; then sh external/oh-my-zsh.sh; fi
