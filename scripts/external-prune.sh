#!/bin/sh
# Remove every version next to <dest> other than <dest> itself.
#   sh scripts/external-prune.sh ~/.local/share/neovim/0.12.6
set -eu

dest=$1
for d in "$(dirname "$dest")"/*; do
  [ -e "$d" ] && [ "$d" != "$dest" ] || continue
  if rm -rf "$d"; then
    echo "remove   $d"
  else
    echo "warning: cannot remove $d" >&2
  fi
done
