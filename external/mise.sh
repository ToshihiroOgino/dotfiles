#!/bin/sh
# Install mise with the official installer. Idempotent.
#   sh external/mise.sh
# Updates are left to `mise self-update`, so an existing install is left alone.
set -eu

if [ -x "$HOME/.local/bin/mise" ] || command -v mise >/dev/null 2>&1; then
  echo "skip     mise (already installed)"
  exit 0
fi

# Saved first so the script is not run half-downloaded.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL -o "$tmp/install.sh" https://mise.run
echo "install  mise"
MISE_INSTALL_HELP=0 sh "$tmp/install.sh"
