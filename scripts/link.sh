#!/bin/sh
# Symlink one source to one target.
#   sh scripts/link.sh [--apply] [--local] <source relative to repo root> <target>
# Existing symlinks are replaced; real files/directories go to <target>.backup.
#
# --local: <source> is a machine-local file (gitignored). When it does not
# exist yet it is created first, from the first existing of
#   <dir>/templates/<name>.<platform>   (platform: windows mac wsl linux)
#   <dir>/templates/<name>
# or as an empty file. An existing local file is never touched.
#
# With --apply every change is also appended, with a timestamp, to log/link.log
# in the repository (the log/ directory is gitignored).
set -eu

# Git Bash copies files on `ln -s` unless real symlinks are requested.
export MSYS=winsymlinks:nativestrict

SCRIPTS=$(cd "$(dirname "$0")" && pwd)
ROOT=$(dirname "$SCRIPTS")
APPLY=0
LOCAL=0
while [ $# -gt 0 ]; do
  case $1 in
  --apply) APPLY=1 ;;
  --local) LOCAL=1 ;;
  *) break ;;
  esac
  shift
done
LOG=$ROOT/log/link.log

# report <message>: print it, and with --apply also append it to the log.
report() {
  echo "$1"
  if [ "$APPLY" -eq 1 ]; then
    mkdir -p "$(dirname "$LOG")"
    printf '%s  %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$1" >>"$LOG"
  fi
}

[ $# -eq 2 ] || { echo "usage: link.sh [--apply] [--local] <source> <target>" >&2; exit 2; }

from=$ROOT/$1
to=$2
# Convert a Windows path (C:\...) to a Git Bash path. No-op elsewhere.
if command -v cygpath >/dev/null 2>&1; then
  to=$(cygpath -u "$2")
fi

# Print the template for the local file $from, or nothing if there is none.
find_template() {
  template_dir=$(dirname "$from")/templates
  template_name=$(basename "$from")
  platform=$(sh "$SCRIPTS/os.sh")
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
    report "create   $1 (from ${template#"$ROOT"/})"
  else
    report "create   $1 (empty)"
  fi
  if [ "$APPLY" -eq 1 ]; then
    mkdir -p "$(dirname "$from")"
    if [ -n "$template" ]; then
      cp "$template" "$from"
    else
      : >"$from"
    fi
  fi
elif [ ! -e "$from" ]; then
  echo "MISSING  $1" >&2
  exit 1
fi

if [ -L "$to" ]; then
  existing=$(readlink "$to")
  if [ "$existing" = "$from" ]; then
    echo "ok       $to"
    exit 0
  fi
  report "relink   $to (was -> $existing)"
elif [ -e "$to" ]; then
  if [ -e "$to.backup" ] || [ -L "$to.backup" ]; then
    echo "error: $to.backup already exists" >&2
    exit 1
  fi
  report "backup   $to -> $to.backup"
else
  report "link     $to"
fi

[ "$APPLY" -eq 1 ] || exit 0
mkdir -p "$(dirname "$to")"
if [ -L "$to" ]; then
  rm "$to"
elif [ -e "$to" ]; then
  mv "$to" "$to.backup"
fi
ln -s "$from" "$to"
