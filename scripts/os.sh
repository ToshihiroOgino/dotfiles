#!/bin/sh
# Print the platform name: windows, mac, wsl or linux.
#   PLATFORM=$(sh scripts/os.sh)
set -eu

case $(uname -s) in
Darwin) echo mac ;;
MINGW* | MSYS* | CYGWIN*) echo windows ;;
Linux)
  if grep -qi microsoft /proc/version 2>/dev/null; then
    echo wsl
  else
    echo linux
  fi
  ;;
*)
  echo "error: unsupported OS: $(uname -s)" >&2
  exit 1
  ;;
esac
