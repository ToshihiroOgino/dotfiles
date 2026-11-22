#!/bin/sh
set -eu

DEST=${GITSTATUS_DIR:-$HOME/.local/share/gitstatus}

if [ -e "$DEST" ] && [ ! -d "$DEST/.git" ]; then
  echo "error: $DEST exists but is not a git clone" >&2
  exit 1
fi
if [ ! -d "$DEST/.git" ]; then
  mkdir -p "$(dirname "$DEST")"
  git clone --depth 1 https://github.com/romkatv/gitstatus.git "$DEST"
  exit 0
fi

if ! git -C "$DEST" fetch --quiet --depth 1 origin HEAD 2>/dev/null; then
  echo "skip     gitstatus (fetch failed)"
  exit 0
fi
if [ "$(git -C "$DEST" rev-parse HEAD)" = "$(git -C "$DEST" rev-parse FETCH_HEAD)" ]; then
  echo "skip     gitstatus (up to date)"
else
  git -C "$DEST" reset --hard --quiet FETCH_HEAD
  echo "update   gitstatus"
fi
