# dotfiles の実行コマンド集。各ディレクトリの Makefile に委譲する。
# Recipes are plain commands so they work under cmd, pwsh, or sh.

# $(MAKE) は空白や括弧を含むパスに展開されうるのでクォートする。
HERDR := "$(MAKE)" -C common/herdr

.PHONY: all herdr windows linux macos check clean imestat

all: herdr

# herdr の設定を全OS分生成する。
herdr:
	$(HERDR)

# OS 名と検証はそのまま common/herdr に渡す。
windows linux macos check clean:
	$(HERDR) $@

imestat:
	"$(MAKE)" -C tools/imestat build
