#!/bin/sh
# Compare pinned versions in external/*.sh with the latest GitHub release. Read-only.
#   sh scripts/external-outdated.sh
# Only tools that do not update themselves are listed.
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)

# name script owner/repo
TOOLS="nvim nvim neovim/neovim
gitleaks gitleaks gitleaks/gitleaks
tree-sitter tree-sitter tree-sitter/tree-sitter"

status=0
echo "$TOOLS" | while read -r name script repo; do
  pinned=$(sed -n 's/^VERSION=//p' "$ROOT/external/$script.sh" | head -n 1)
  # The /releases/latest redirect needs no API token and has no rate limit.
  if ! url=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest"); then
    printf '%-12s %-10s ?  (fetch failed)\n' "$name" "$pinned"
    continue
  fi
  latest=${url##*/}
  latest=${latest#v}
  if [ "$pinned" = "$latest" ]; then
    printf '%-12s %-10s up to date\n' "$name" "$pinned"
  else
    printf '%-12s %-10s -> %s\n' "$name" "$pinned" "$latest"
  fi
done
exit "$status"
