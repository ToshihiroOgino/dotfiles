# Shell and platform for every target, included from the root Makefile.
# On Windows the recipes run in Git for Windows' sh, so the scripts work as
# on unix. Defines PLATFORM (windows mac wsl linux) and FAMILY (windows unix).

# Directory of this file, so os.sh is found from any working directory.
ENV_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(OS),Windows_NT)
# Not abspath: it splits on spaces, which breaks C:/Program Files/Git.
GIT_ROOT := $(subst \,/,$(shell git --exec-path))
$(foreach e,mingw64 mingw32 ucrt64 clang64 clangarm64,$(eval GIT_ROOT := $$(patsubst %/$e/libexec/git-core,%,$$(GIT_ROOT))))
# wildcard needs the spaces escaped too.
GIT_SH := $(GIT_ROOT)/usr/bin/sh.exe
$(if $(wildcard $(subst $() ,\ ,$(GIT_SH))),,$(error Git for Windows not found: $(GIT_SH)))
export PATH := $(GIT_ROOT)/usr/bin;$(PATH)
SHELL := $(GIT_ROOT)/usr/bin/sh.exe
# pwsh/cmd do not define HOME.
HOME ?= $(subst \,/,$(USERPROFILE))
export HOME
endif

PLATFORM := $(shell sh $(ENV_DIR)os.sh)
$(if $(PLATFORM),,$(error cannot detect the platform: $(ENV_DIR)os.sh failed))
FAMILY := $(if $(filter windows,$(PLATFORM)),windows,unix)
