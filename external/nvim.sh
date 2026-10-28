#!/bin/sh
# Install a pinned Neovim release from the official tarball. Idempotent.
#   sh external/nvim.sh
# Extracted to ~/.local/share/neovim/<version> (not share/nvim: that is nvim's
# own data dir); on unix ~/.local/bin/nvim links to it, older versions are kept.
# Windows has no link: its bin goes on the user PATH (external/path-windows.ps1).
set -eu

VERSION=0.12.6
ROOT=$(cd "$(dirname "$0")/.." && pwd)
PLATFORM=$(sh "$ROOT/scripts/os.sh")
DEST=$HOME/.local/share/neovim/$VERSION
BIN=$HOME/.local/bin

case $PLATFORM-$(uname -m) in
wsl-x86_64 | linux-x86_64) asset=nvim-linux-x86_64.tar.gz ;;
wsl-aarch64 | linux-aarch64) asset=nvim-linux-arm64.tar.gz ;;
mac-arm64) asset=nvim-macos-arm64.tar.gz ;;
mac-x86_64) asset=nvim-macos-x86_64.tar.gz ;;
windows-*) asset=nvim-win64.zip ;;
*)
  echo "error: unsupported platform: $PLATFORM-$(uname -m)" >&2
  exit 1
  ;;
esac

exe=$DEST/bin/nvim
[ "$PLATFORM" = windows ] && exe=$exe.exe

# An nvim on PATH that is not one of ours would shadow or be shadowed by ours.
if found=$(command -v nvim 2>/dev/null); then
  root=$(dirname "$DEST")
  case $(readlink -f "$found") in
  "$(readlink -f "$root" 2>/dev/null || echo "$root")"/*) ;;
  *)
    echo "error: another nvim is on PATH: $found" >&2
    exit 1
    ;;
  esac
fi

if [ -x "$exe" ] && [ "$("$exe" --version | head -n 1)" = "NVIM v$VERSION" ]; then
  echo "skip     nvim v$VERSION (already installed)"
else
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  echo "download $asset (v$VERSION)"
  curl -fsSL -o "$tmp/$asset" "https://github.com/neovim/neovim/releases/download/v$VERSION/$asset"
  mkdir "$tmp/x"
  case $asset in
  *.zip) powershell.exe -NoProfile -Command "Expand-Archive -LiteralPath '$(cygpath -w "$tmp/$asset")' -DestinationPath '$(cygpath -w "$tmp/x")'" ;;
  *) tar -xzf "$tmp/$asset" -C "$tmp/x" ;;
  esac
  # The archive holds a single top-level directory.
  rm -rf "$DEST"
  mkdir -p "$(dirname "$DEST")"
  mv"$tmp"/x/* "$DEST"
  echo "install  $DEST (v$VERSION)"
fi

if [ "$PLATFORM" = windows ]; then
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(cygpath -w "$ROOT/external/path-windows.ps1")" \
    -Dir "$(cygpath -w "$DEST/bin")" -Root "$(cygpath -w "$(dirname "$DEST")")"
else
  mkdir -p "$BIN"
  if [ "$(readlink "$BIN/nvim" 2>/dev/null)" != "$DEST/bin/nvim" ]; then
    ln -sfn "$DEST/bin/nvim" "$BIN/nvim"
    echo "link     $BIN/nvim -> $DEST/bin/nvim"
  fi
fi
