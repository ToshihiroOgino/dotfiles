#!/bin/sh
# Launch ghostty from wslg.exe. Go through zsh so ~/.zshenv (GDK_BACKEND, fcitx, ...) is read.
# On a cold WSL start, a window created before the WSLg RDP client connects never shows,
# and the single-instance ghostty then serves every later launch with that dead window.
# Wait for the handshake (give up after 15s).
i=0
until grep -q "Client HandShake" /mnt/wslg/weston.log 2>/dev/null || [ "$i" -ge 150 ]; do
	sleep 0.1
	i=$((i + 1))
done
exec zsh -c 'exec /usr/bin/ghostty "$@"' ghostty "$@"