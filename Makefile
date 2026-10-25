# dotfiles commands. The recipes need sh (Git Bash on Windows).

.DEFAULT_GOAL := link

include scripts/link.mk

all: link imestat pkgsync systemd-enable

.PHONY: all imestat pkgsync systemd-enable

imestat:
	"$(MAKE)" -C tools/imestat build

pkgsync:
	"$(MAKE)" -C tools/pkgsync build

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh
