#!/bin/sh
# Install Claude Code with the official installer. Idempotent.
#   sh external/claude.sh
# Claude Code updates itself afterwards, so an existing install is left alone.
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
PLATFORM=$(sh "$ROOT/scripts/os.sh")

if [ "$PLATFORM" = windows ]; then
  if command -v claude >/dev/null 2>&1 || [ -e "$HOME/.local/bin/claude.exe" ]; then
    echo "skip     claude (already installed)"
    exit 0
  fi
  echo "install  claude"
  powershell.exe -NoProfile -Command 'irm https://claude.ai/install.ps1 | iex'
  exit 0
fi

if [ -x "$HOME/.local/bin/claude" ] || command -v claude >/dev/null 2>&1; then
  echo "skip     claude (already installed)"
  exit 0
fi

# Saved first so the script is not run half-downloaded.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL -o "$tmp/install.sh" https://claude.ai/install.sh
echo "install  claude"
bash "$tmp/install.sh"
