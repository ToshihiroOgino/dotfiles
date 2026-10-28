#!/bin/sh
# Install a pinned tree-sitter CLI from the official release. Idempotent.
#   sh external/tree-sitter.sh
# nvim-treesitter (main) needs 0.26.1 or later, newer than most distro packages.
# Extracted to ~/.local/share/tree-sitter/<version>; on unix ~/.local/bin/tree-sitter
# links to it, older versions are kept. Windows has no link: its dir goes on the
# user PATH (external/path-windows.ps1).
set -eu

VERSION=0.27.1
ROOT=$(cd "$(dirname "$0")/.." && pwd)
PLATFORM=$(sh "$ROOT/scripts/os.sh")
DEST=$HOME/.local/share/tree-sitter/$VERSION
BIN=$HOME/.local/bin

case $PLATFORM-$(uname -m) in
wsl-x86_64 | linux-x86_64) asset=tree-sitter-linux-x64.gz ;;
wsl-aarch64 | linux-aarch64) asset=tree-sitter-linux-arm64.gz ;;
mac-arm64) asset=tree-sitter-macos-arm64.gz ;;
mac-x86_64) asset=tree-sitter-macos-x64.gz ;;
windows-*) asset=tree-sitter-windows-x64.gz ;;
*)
  echo "error: unsupported platform: $PLATFORM-$(uname -m)" >&2
  exit 1
  ;;
esac

exe=$DEST/tree-sitter
[ "$PLATFORM" = windows ] && exe=$exe.exe

if [ -x "$exe" ] && [ "$("$exe" --version | cut -d' ' -f1,2)" = "tree-sitter $VERSION" ]; then
  echo "skip     tree-sitter v$VERSION (already installed)"
else
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  echo "download $asset (v$VERSION)"
  curl -fsSL -o "$tmp/$asset" "https://github.com/tree-sitter/tree-sitter/releases/download/v$VERSION/$asset"
  # The release holds the bare binary, gzipped.
  gzip -dc "$tmp/$asset" >"$tmp/tree-sitter"
  chmod +x "$tmp/tree-sitter"
  mkdir -p "$DEST"
  mv "$tmp/tree-sitter" "$exe"
  echo "install  $DEST (v$VERSION)"
fi

if [ "$PLATFORM" = windows ]; then
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(cygpath -w "$ROOT/external/path-windows.ps1")" \
    -Dir "$(cygpath -w "$DEST")" -Root "$(cygpath -w "$(dirname "$DEST")")"
else
  mkdir -p "$BIN"
  if [ "$(readlink "$BIN/tree-sitter" 2>/dev/null)" != "$exe" ]; then
    ln -sfn "$exe" "$BIN/tree-sitter"
    echo "link     $BIN/tree-sitter -> $exe"
  fi
fi
