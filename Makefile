# dotfiles commands. The recipes need sh (Git Bash on Windows).

.DEFAULT_GOAL := link

include scripts/link.mk

.PHONY: imestat systemd-enable

imestat:
	"$(MAKE)" -C tools/imestat build

# Not part of `make link`: enabling a unit changes systemd state. WSL only.
SYSTEMD_UNITS := fcitx5.service
systemd-enable:
	@if [ "$(PLATFORM)" = wsl ] && command -v systemctl >/dev/null 2>&1; then \
		systemctl --user daemon-reload; \
		systemctl --user enable $(SYSTEMD_UNITS); \
	else \
		echo "skip     systemd-enable (WSL with systemctl only)"; \
	fi
