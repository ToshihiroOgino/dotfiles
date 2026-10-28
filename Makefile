.DEFAULT_GOAL := link

include scripts/link.mk

all: tools link systemd-enable

.PHONY: all imestat pkgsync systemd-enable

tools: imestat pkgsync

imestat:
	"$(MAKE)" -C tools/imestat build

pkgsync:
	"$(MAKE)" -C tools/pkgsync build

systemd-enable:
	@sh $(SCRIPTS)systemd-enable.sh
