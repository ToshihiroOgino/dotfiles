#!/bin/sh
# Shortcut command:
#   "C:\Program Files\WSL\wslg.exe" -d Ubuntu-26.04 --cd "~" -- /home/tshr/dotfiles/scripts/ghostty-wslg.sh
i=0
until { grep -q "Client HandShake" /mnt/wslg/weston.log 2>/dev/null \
	&& systemctl --user is-active --quiet fcitx5.service 2>/dev/null; } || [ "$i" -ge 150 ]; do
        sleep 0.1
        i=$((i + 1))
done
exec zsh -c 'exec /usr/bin/ghostty "$@"' ghostty "$@"
