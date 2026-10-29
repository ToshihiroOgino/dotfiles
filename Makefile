.DEFAULT_GOAL := link

include scripts/env.mk
include scripts/link.mk

all: tools link external systemd-enable

.PHONY: all ime pkgsync systemd-enable external nvim tree-sitter omz claude tpm

# Built tools are linked into BIN, which is on PATH.
BIN := $(HOME)/.local/bin
EXE := $(if $(filter windows,$(FAMILY)),.exe)

tools: ime pkgsync

ime:
	"$(MAKE)" -C tools/ime build
	$(LINK) tools/ime/ime$(EXE) "$(BIN)/ime$(EXE)"

pkgsync:
	"$(MAKE)" -C tools/pkgsync build
	$(LINK) tools/pkgsync/pkgsync$(EXE) "$(BIN)/pkgsync$(EXE)"

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh

external: nvim tree-sitter omz claude tpm

nvim:
	@sh external/nvim.sh

tree-sitter:
	@sh external/tree-sitter.sh

claude:
	@sh external/claude.sh

omz:
	@if [ "$(PLATFORM)" != windows ]; then sh external/oh-my-zsh.sh; fi

tpm:
	@if [ "$(PLATFORM)" != windows ]; then sh external/tpm.sh; fi
