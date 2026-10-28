# Shell and platform for every target, included from the root Makefile.
# On Windows the recipes run in Git for Windows' sh, so the scripts work as
# on unix. Defines PLATFORM (windows mac wsl linux) and FAMILY (windows unix).

# Directory of this file, so os.sh is found from any working directory.
ENV_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(OS),Windows_NT)
GIT_ROOT := $(abspath $(shell git --exec-path)/../../..)
$(if $(wildcard $(GIT_ROOT)/usr/bin/sh.exe),,$(error Git for Windows not found: $(GIT_ROOT)/usr/bin/sh.exe))
export PATH := $(GIT_ROOT)/usr/bin;$(PATH)
SHELL := $(GIT_ROOT)/usr/bin/sh.exe
# pwsh/cmd do not define HOME.
HOME ?= $(subst \,/,$(USERPROFILE))
export HOME
endif

PLATFORM := $(shell sh $(ENV_DIR)os.sh)
$(if $(PLATFORM),,$(error cannot detect the platform: $(ENV_DIR)os.sh failed))
FAMILY := $(if $(filter windows,$(PLATFORM)),windows,unix)
