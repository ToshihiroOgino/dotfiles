#!/bin/sh
# Install a pinned gitleaks release. Idempotent.
#   sh external/gitleaks.sh
# apt ships a much older version.
set -eu

VERSION=8.30.1
ROOT=$(cd "$(dirname "$0")/.." && pwd)
PLATFORM=$(sh "$ROOT/scripts/os.sh")
DEST=$HOME/.local/share/gitleaks/$VERSION
BIN=$HOME/.local/bin
URL=https://github.com/gitleaks/gitleaks/releases/download/v$VERSION

case $PLATFORM-$(uname -m) in
wsl-x86_64 | linux-x86_64) asset=gitleaks_${VERSION}_linux_x64.tar.gz ;;
wsl-aarch64 | linux-aarch64) asset=gitleaks_${VERSION}_linux_arm64.tar.gz ;;
mac-arm64) asset=gitleaks_${VERSION}_darwin_arm64.tar.gz ;;
mac-x86_64) asset=gitleaks_${VERSION}_darwin_x64.tar.gz ;;
windows-*) asset=gitleaks_${VERSION}_windows_x64.zip ;;
*)
  echo "error: unsupported platform: $PLATFORM-$(uname -m)" >&2
  exit 1
  ;;
esac

name=gitleaks
[ "$PLATFORM" = windows ] && name=gitleaks.exe
exe=$DEST/$name

# A gitleaks on PATH that is not one of ours would shadow or be shadowed by ours.
if found=$(command -v gitleaks 2>/dev/null); then
  root=$(dirname "$DEST")
  case $(readlink -f "$found") in
  "$(readlink -f "$root" 2>/dev/null || echo "$root")"/*) ;;
  *)
    echo "error: another gitleaks is on PATH: $found" >&2
    exit 1
    ;;
  esac
fi

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

if [ -x "$exe" ] && [ "$("$exe" version)" = "$VERSION" ]; then
  echo "skip     gitleaks v$VERSION (already installed)"
else
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  echo "download $asset (v$VERSION)"
  curl -fsSL --retry 3 --connect-timeout 10 -o "$tmp/$asset" "$URL/$asset"
  curl -fsSL --retry 3 --connect-timeout 10 -o "$tmp/checksums.txt" "$URL/gitleaks_${VERSION}_checksums.txt"
  expected=$(awk -v f="$asset" '$2 == f { print $1 }' "$tmp/checksums.txt")
  if [ -z "$expected" ] || [ "$expected" != "$(sha256 "$tmp/$asset")" ]; then
    echo "error: checksum mismatch: $asset" >&2
    exit 1
  fi
  mkdir "$tmp/x"
  case $asset in
  *.zip)
    ZIP=$(cygpath -w "$tmp/$asset") OUT=$(cygpath -w "$tmp/x") \
      powershell.exe -NoProfile -Command 'Expand-Archive -LiteralPath $env:ZIP -DestinationPath $env:OUT'
    ;;
  *) tar -xzf "$tmp/$asset" -C "$tmp/x" ;;
  esac
  chmod +x "$tmp/x/$name"
  mkdir -p "$DEST"
  mv "$tmp/x/$name" "$exe"
  echo "install  $DEST (v$VERSION)"
fi

if [ "$PLATFORM" = windows ]; then
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(cygpath -w "$ROOT/external/path-windows.ps1")" \
    -Dir "$(cygpath -w "$DEST")" -Root "$(cygpath -w "$(dirname "$DEST")")"
else
  mkdir -p "$BIN"
  if [ "$(readlink "$BIN/gitleaks" 2>/dev/null)" != "$exe" ]; then
    ln -sfn "$exe" "$BIN/gitleaks"
    echo "link     $BIN/gitleaks -> $exe"
  fi
fi
