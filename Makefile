.DEFAULT_GOAL := link

include scripts/link.mk

all: link ime pkgsync systemd-enable

.PHONY: all ime pkgsync systemd-enable

ime:
	"$(MAKE)" -C tools/ime build

pkgsync:
	"$(MAKE)" -C tools/pkgsync build

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh
