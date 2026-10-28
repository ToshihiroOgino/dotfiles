.DEFAULT_GOAL := link

include scripts/link.mk

all: tools link systemd-enable

.PHONY: all ime pkgsync systemd-enable

tools: ime pkgsync

ime:
	"$(MAKE)" -C tools/ime build

pkgsync:
	"$(MAKE)" -C tools/pkgsync build

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh
