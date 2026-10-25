#!/bin/sh
# Enable every unit in config/wsl/systemd. WSL only; skipped elsewhere.
#   sh scripts/systemd-enable.sh
set -eu

dir=$(dirname "$0")

if [ "$(sh "$dir/os.sh")" != wsl ]; then
	echo "skip     systemd-enable (WSL only)"
	exit 0
fi

units=
for f in "$dir"/../config/wsl/systemd/*.service; do
	units="$units $(basename "$f")"
done

systemctl --user daemon-reload
# shellcheck disable=SC2086
systemctl --user enable $units
