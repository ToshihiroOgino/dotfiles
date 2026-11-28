#!/bin/sh
set -eu

PLUGINS=$HOME/.psmux/plugins
DEST=$PLUGINS/ppm

if [ -d "$DEST" ]; then
  echo "skip     ppm (already installed: $DEST)"
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

git clone --depth 1 https://github.com/psmux/psmux-plugins.git "$tmp"
mkdir -p "$PLUGINS"
cp -r "$tmp/ppm" "$DEST"
echo "install  ppm ($DEST)"
