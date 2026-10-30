#!/bin/sh
# Shallow-clone tpm (tmux plugin manager) into ~/.tmux/plugins/tpm, then install
# the plugins listed in tmux.conf. Idempotent.
#   sh external/tpm.sh
# Needs a tmux server to read @plugin from: a running one is reused, otherwise a
# detached one is started for the install and killed afterwards.
set -eu

PLUGINS=${TMUX_PLUGIN_MANAGER_PATH:-$HOME/.tmux/plugins}
DEST=$PLUGINS/tpm

if [ -d "$DEST/.git" ]; then
  echo "skip     tpm (already cloned: $DEST)"
elif [ -e "$DEST" ]; then
  echo "error: $DEST exists but is not a git clone" >&2
  exit 1
else
  git clone --depth 1 https://github.com/tmux-plugins/tpm.git "$DEST"
fi

if ! command -v tmux >/dev/null 2>&1; then
  echo "skip     tmux plugins (tmux not installed)"
  exit 0
fi

started=
if ! tmux has-session 2>/dev/null; then
  tmux new-session -d -s tpm-install
  started=1
fi
trap '[ -z "$started" ] || tmux kill-server' EXIT

tmux set-environment -g TMUX_PLUGIN_MANAGER_PATH "$PLUGINS/"

"$DEST/bin/install_plugins"
