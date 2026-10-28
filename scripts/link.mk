# Symlink definitions, included from the root Makefile.
#   make link              # all tools for this OS
#   make link-nvim         # one tool

# Directory of this file, so the scripts are found from any working directory.
SCRIPTS := $(dir $(lastword $(MAKEFILE_LIST)))

LINK := sh $(SCRIPTS)link.sh
# For machine-local files: created from templates/ when missing.
LINK_LOCAL := $(LINK) --local

COMMON        := git claude tmux nvim nvim-tshr vim lazygit alacritty herdr
TOOLS_wsl     := $(COMMON) ghostty wsl zsh
TOOLS_linux   := $(COMMON) ghostty zsh
TOOLS_mac     := $(COMMON) ghostty zsh
TOOLS_windows := $(COMMON) pwsh

ALL_TOOLS := $(sort $(TOOLS_wsl) $(TOOLS_linux) $(TOOLS_mac) $(TOOLS_windows))

.PHONY: link $(addprefix link-,$(ALL_TOOLS))

link: $(addprefix link-,$(TOOLS_$(PLATFORM)))

link-git:
	$(LINK) config/git "$(HOME)/.config/git"
	@[ -e "$(HOME)/.gitconfig" ] || { : >"$(HOME)/.gitconfig" && echo "create   $(HOME)/.gitconfig (empty)"; }
link-claude:
	$(LINK) config/claude "$(HOME)/.claude"
link-tmux:
	$(LINK) config/tmux/tmux.conf "$(HOME)/.tmux.conf"
	$(LINK_LOCAL) config/tmux/tmux.local.conf "$(HOME)/.tmux.local.conf"

NVIM_unix    := $(HOME)/.config/nvim
NVIM_windows := $(LOCALAPPDATA)/nvim
link-nvim:
	$(LINK) config/nvim "$(NVIM_$(FAMILY))"

# NVIM_APPNAME=nvim-tshr (alias `v`): the from-scratch config that will replace nvim.
NVIM_TSHR_unix    := $(HOME)/.config/nvim-tshr
NVIM_TSHR_windows := $(LOCALAPPDATA)/nvim-tshr
link-nvim-tshr:
	$(LINK) config/nvim-tshr "$(NVIM_TSHR_$(FAMILY))"

link-zsh:
	$(LINK) config/zsh/zshrc "$(HOME)/.zshrc"
	$(LINK_LOCAL) config/zsh/zshrc.local "$(HOME)/.local.zshrc"

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

# WSLg-only settings: input method (fcitx5), GPU and GTK environment.
# Run `make systemd-enable` once to start fcitx5.service at login.
link-wsl:
	$(LINK) config/wsl/zshenv "$(HOME)/.zshenv"
	$(LINK) config/wsl/systemd/fcitx5.service "$(HOME)/.config/systemd/user/fcitx5.service"

link-pwsh:
	@sh $(SCRIPTS)link-pwsh.sh
