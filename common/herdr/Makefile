# herdr の設定生成。
# Recipes are plain commands so they work under cmd, pwsh, or sh.

# Lua 5.1 以上。lua が無い環境では `make LUA="nvim -l"` で代用できる。
LUA ?= lua
GEN := $(LUA) gen-config.lua

.PHONY: all gen windows linux macos check clean

all: gen

# 全OS分を generated/ に生成する。
gen:
	$(GEN)

windows:
	$(GEN) windows

linux:
	$(GEN) linux

macos:
	$(GEN) macos

# 設定ディレクトリに貼った symlink 経由で、今の環境の設定を検証する。
check:
	herdr config check

clean:
	$(GEN) --clean
