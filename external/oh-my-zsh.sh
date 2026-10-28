#!/bin/sh
# Shallow-clone oh-my-zsh into ~/.oh-my-zsh. Idempotent; never touches ~/.zshrc.
#   sh external/oh-my-zsh.sh
# Updates are left to `omz update` (zstyle ':omz:update' in config/zsh/zshrc).
set -eu

DEST=${ZSH:-$HOME/.oh-my-zsh}

if [ -d "$DEST/.git" ]; then
  echo "skip     oh-my-zsh (already cloned: $DEST)"
  exit 0
fi
if [ -e "$DEST" ]; then
  echo "error: $DEST exists but is not a git clone" >&2
  exit 1
fi

git clone --depth 1 https://github.com/ohmyzsh/ohmyzsh.git "$DEST"
