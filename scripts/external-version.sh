#!/bin/sh
# Print the latest stable release version of a GitHub repo, without the leading v.
# Offline, fall back to the newest version already installed under <store>.
#   VERSION=$(sh scripts/external-version.sh neovim/neovim ~/.local/share/neovim)
set -eu

repo=$1
store=$2

if url=$(curl -fsSL --retry 2 --connect-timeout 10 -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest"); then
  version=${url##*/}
  version=${version#v}
  if echo "$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "$version"
    exit 0
  fi
  echo "warning: unexpected release tag for $repo: $version" >&2
else
  echo "warning: cannot fetch the latest release of $repo" >&2
fi

installed=$(ls -1 "$store" 2>/dev/null | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1)
if [ -z "$installed" ]; then
  echo "error: no installed version of $repo to fall back to" >&2
  exit 1
fi
echo "warning: using installed $installed" >&2
echo "$installed"
