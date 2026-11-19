#!/bin/sh
# UPDATE=1 resets outdated plugins to the latest; local changes are lost.
set -eu

OMZ=${ZSH:-$HOME/.oh-my-zsh}
DEST=${ZSH_CUSTOM:-$OMZ/custom}/plugins

PLUGINS="
zsh-autosuggestions https://github.com/zsh-users/zsh-autosuggestions.git
"

if [ ! -d "$OMZ" ]; then
  echo "skip     zsh-plugins (oh-my-zsh not installed)"
  exit 0
fi
mkdir -p "$DEST"

echo "$PLUGINS" | while read -r name url; do
  [ -n "$name" ] || continue
  dir=$DEST/$name
  if [ -e "$dir" ] && [ ! -d "$dir/.git" ]; then
    echo "error: $dir exists but is not a git clone" >&2
    exit 1
  fi
  if [ ! -d "$dir/.git" ]; then
    git clone --depth 1 "$url" "$dir"
    continue
  fi
  if ! git -C "$dir" fetch --quiet --depth 1 origin HEAD 2>/dev/null; then
    echo "skip     $name (fetch failed)"
    continue
  fi
  local_rev=$(git -C "$dir" rev-parse HEAD)
  remote_rev=$(git -C "$dir" rev-parse FETCH_HEAD)
  if [ "$local_rev" = "$remote_rev" ]; then
    echo "skip     $name (up to date)"
  elif [ "${UPDATE:-}" = 1 ]; then
    git -C "$dir" reset --hard --quiet FETCH_HEAD
    echo "update   $name"
  else
    echo "outdated $name (UPDATE=1 make zsh-plugins)"
  fi
done
