# Symlink definitions, included from the root Makefile.
#   make link              # dry-run, all tools for this OS
#   make link APPLY=1      # APPLY must be exactly 1
#   make link-nvim         # one tool

# Directory of this file, so the scripts are found from any working directory.
SCRIPTS := $(dir $(lastword $(MAKEFILE_LIST)))

# ---- OS detection ----
PLATFORM := $(shell sh $(SCRIPTS)os.sh)
$(if $(PLATFORM),,$(error cannot detect the platform: $(SCRIPTS)os.sh failed))
FAMILY := $(if $(filter windows,$(PLATFORM)),windows,unix)

LINK := sh $(SCRIPTS)link.sh $(if $(filter 1,$(APPLY)),--apply)
# For machine-local files: created from templates/ when missing.
LINK_LOCAL := $(LINK) --local

# ---- which tools on which OS ----
COMMON        := git claude tmux nvim vim lazygit alacritty herdr
TOOLS_wsl     := $(COMMON) ghostty
TOOLS_linux   := $(COMMON) ghostty
TOOLS_mac     := $(COMMON) ghostty
TOOLS_windows := $(COMMON) pwsh

ALL_TOOLS := $(sort $(TOOLS_wsl) $(TOOLS_linux) $(TOOLS_mac) $(TOOLS_windows))

.PHONY: link $(addprefix link-,$(ALL_TOOLS))

link: $(addprefix link-,$(TOOLS_$(PLATFORM)))

# ---- where each tool goes (per OS family when it differs) ----
link-git:
	$(LINK) config/git "$(HOME)/.config/git"
link-claude:
	$(LINK) config/claude "$(HOME)/.claude"
link-tmux:
	$(LINK) config/tmux/tmux.conf "$(HOME)/.tmux.conf"
	$(LINK_LOCAL) config/tmux/tmux.local.conf "$(HOME)/.tmux.local.conf"

NVIM_unix    := $(HOME)/.config/nvim
NVIM_windows := $(LOCALAPPDATA)/nvim
link-nvim:
	$(LINK) config/nvim "$(NVIM_$(FAMILY))"

link-vim:
	$(LINK) config/vim/vimrc "$(HOME)/.vimrc"
	$(LINK_LOCAL) config/vim/vimrc.local "$(HOME)/.vimrc.local"

LAZYGIT_unix    := $(HOME)/.config/lazygit/config.yml
LAZYGIT_windows := $(APPDATA)/lazygit/config.yml
link-lazygit:
	$(LINK) config/lazygit/config.yml "$(LAZYGIT_$(FAMILY))"

HERDR_DST_unix    := $(HOME)/.config/herdr/config.toml
HERDR_DST_windows := $(APPDATA)/herdr/config.toml
link-herdr:
	$(LINK) config/herdr/config.toml "$(HERDR_DST_$(FAMILY))"

ALACRITTY_DIR_unix    := $(HOME)/.config/alacritty
ALACRITTY_DIR_windows := $(APPDATA)/alacritty
link-alacritty:
	$(LINK) config/alacritty/alacritty.toml "$(ALACRITTY_DIR_$(FAMILY))/alacritty.toml"
	$(LINK_LOCAL) config/alacritty/alacritty.local.toml "$(ALACRITTY_DIR_$(FAMILY))/alacritty.local.toml"

link-ghostty:
	$(LINK) config/ghostty/config.ghostty "$(HOME)/.config/ghostty/config.ghostty"
	$(LINK_LOCAL) config/ghostty/config.local "$(HOME)/.config/ghostty/config.local"

# Ask pwsh where its profile goes; skip when pwsh is not installed.
# link.sh converts the Windows path it returns.
link-pwsh:
	@if command -v pwsh >/dev/null 2>&1; then \
		profile=$$(pwsh -NoProfile -Command '$$PROFILE.CurrentUserAllHosts' | tr -d '\r'); \
		[ -n "$$profile" ] || { echo "error: pwsh did not return the profile path" >&2; exit 1; }; \
		$(LINK) config/pwsh/profile.ps1 "$$profile"; \
		$(LINK_LOCAL) config/pwsh/profile.local.ps1 "$$(dirname "$$(cygpath -u "$$profile")")/profile.local.ps1"; \
	else \
		echo "skip     pwsh (not installed)"; \
	fi
