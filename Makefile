.DEFAULT_GOAL := link

include scripts/env.mk
include scripts/link.mk

all: tools link external systemd-enable

.PHONY: all pkgsync claude-tmux-state gpvpn systemd-enable external nvim tree-sitter gitleaks omz omz-plugins gitstatus claude mise tpm ppm

EXE := $(if $(filter windows,$(FAMILY)),.exe)

tools: pkgsync claude-tmux-state $(if $(filter windows,$(FAMILY)),,gpvpn)

pkgsync:
	"$(MAKE)" -C tools/pkgsync build
	$(LINK) tools/pkgsync/pkgsync$(EXE) "$(BIN)/pkgsync$(EXE)"

claude-tmux-state:
	"$(MAKE)" -C tools/claude-tmux-state build
	$(LINK) tools/claude-tmux-state/claude-tmux-state$(EXE) "$(BIN)/claude-tmux-state$(EXE)"

gpvpn:
	$(LINK) tools/gpvpn/gpvpn "$(BIN)/gpvpn"
	@[ -e tools/gpvpn/gpvpn.conf ] || { cp tools/gpvpn/templates/gpvpn.conf tools/gpvpn/gpvpn.conf && echo "create   tools/gpvpn/gpvpn.conf (from template)"; }

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh

external: nvim tree-sitter gitleaks omz omz-plugins gitstatus claude mise tpm ppm

nvim:
	@sh external/nvim.sh

gitleaks:
	@sh external/gitleaks.sh

tree-sitter:
	@sh external/tree-sitter.sh

gitstatus:
	@if [ "$(PLATFORM)" != windows ]; then sh external/gitstatus.sh; fi

claude:
	@sh external/claude.sh

mise:
	@if [ "$(PLATFORM)" != windows ]; then sh external/mise.sh; fi

omz:
	@if [ "$(PLATFORM)" != windows ]; then sh external/oh-my-zsh.sh; fi

omz-plugins: omz
	@if [ "$(PLATFORM)" != windows ]; then sh external/omz-plugins.sh; fi

tpm:
	@if [ "$(PLATFORM)" != windows ]; then sh external/tpm.sh; fi

ppm:
	@if [ "$(PLATFORM)" = windows ]; then sh external/ppm.sh; fi
