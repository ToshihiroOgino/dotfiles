#!/bin/sh
# Symlink one source to one target.
#   sh scripts/link.sh [--local] <source relative to repo root> <target>
# Existing symlinks are replaced; real files/directories go to <target>.backup.
#
# --local: <source> is a machine-local file (gitignored). When it does not
# exist yet it is created first, from the first existing of
#   <dir>/templates/<name>.<platform>   (platform: windows mac wsl linux)
#   <dir>/templates/<name>
# or as an empty file. An existing local file is never touched.
#
# Every change is appended, with a timestamp, to log/link.log (gitignored).
set -eu

# Git Bash copies files on `ln -s` unless real symlinks are requested.
export MSYS=winsymlinks:nativestrict

ROOT=$(cd "$(dirname "$0")/.." && pwd)
LOG=$ROOT/log/link.log
LOCAL=0
# Take --local out of the arguments, wherever it is.
for argument in "$@"; do
  shift
  if [ "$argument" = "--local" ]; then
    LOCAL=1
  else
    set -- "$@" "$argument"
  fi
done
[ $# -eq 2 ] || {
  echo "usage: link.sh [--local] <source> <target>" >&2
  exit 2
}

from=$ROOT/$1
to=$2
# Convert a Windows path (C:\...) to a Git Bash path. No-op elsewhere.
if command -v cygpath >/dev/null 2>&1; then
  to=$(cygpath -u "$2")
fi

fail() {
  echo "error    $1" >&2
  exit 1
}

report() {
  echo "$1"
  mkdir -p "$(dirname "$LOG")"
  printf '%s  %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$1" >>"$LOG"
}

find_template() {
  template_dir=$(dirname "$from")/templates
  template_name=$(basename "$from")
  platform=$(sh "$ROOT/scripts/os.sh")
  for candidate in "$template_dir/$template_name.$platform" "$template_dir/$template_name"; do
    if [ -f "$candidate" ]; then
      echo "$candidate"
      return 0
    fi
  done
}

if [ "$LOCAL" -eq 1 ] && [ ! -e "$from" ]; then
  template=$(find_template)
  if [ -n "$template" ]; then
    cp "$template" "$from"
    report "create   $1 (from ${template#"$ROOT"/})"
  else
    : >"$from"
    report "create   $1 (empty)"
  fi
elif [ ! -e "$from" ]; then
  fail "$1 does not exist"
fi

if [ -L "$to" ]; then
  existing=$(readlink "$to")
  if [ "$existing" = "$from" ]; then
    echo "ok       $to"
    exit 0
  fi
  message="relink   $to (was -> $existing)"
elif [ -e "$to" ]; then
  if [ -e "$to.backup" ] || [ -L "$to.backup" ]; then
    fail "$to.backup already exists"
  fi
  mv "$to" "$to.backup"
  report "backup   $to -> $to.backup"
  message="link     $to"
else
  mkdir -p "$(dirname "$to")"
  message="link     $to"
fi
ln -sfn "$from" "$to"
report "$message"
