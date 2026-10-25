#!/bin/sh
# Link the pwsh profile to wherever pwsh looks for it.
#   sh scripts/link-pwsh.sh
# Skipped when pwsh is not installed. link.sh converts the Windows path.
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)

if ! command -v pwsh >/dev/null 2>&1; then
  echo "skip     pwsh (not installed)"
  exit 0
fi

profile=$(pwsh -NoProfile -Command '$PROFILE.CurrentUserAllHosts' | tr -d '\r')
[ -n "$profile" ] || {
  echo "error: pwsh did not return the profile path" >&2
  exit 1
}

profile_dir=$(dirname "$(cygpath -u "$profile")")

sh "$ROOT/scripts/link.sh" config/pwsh/profile.ps1 "$profile"
sh "$ROOT/scripts/link.sh" --local config/pwsh/profile.local.ps1 "$profile_dir/profile.local.ps1"
