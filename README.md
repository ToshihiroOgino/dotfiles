# Dotfiles

## Usage

Clone this repo into ~/dotfiles.

## Make

ルートの Makefile は各ディレクトリの Makefile に委譲する。

```sh
make            # herdr の設定を全OS分生成する
make windows    # herdr の設定を Windows 分だけ生成する (linux / macos も同様)
make check      # herdr の設定を検証する
make imestat    # tools/imestat をビルドする
```

## Claude Code

`~/.claude` ごと `common/claude` へリンクして管理する。実行時データ(履歴・認証情報など)は
`common/claude/.gitignore` のホワイトリストで除外している。

```powershell
# Windows (Claude Code を全て終了してから、~/.claude を common/claude へ移動して張り直す)
New-Item -ItemType SymbolicLink -Path ~\.claude -Target ~\dotfiles\common\claude
```

```sh
# Linux / macOS
ln -s ~/dotfiles/common/claude ~/.claude
```

既存の `~/.claude` がある場合は、先に中身を `common/claude` へ移動しておく。

- 管理対象: `CLAUDE.md` / `keybindings.json` / `settings.json` / `skills/` (自作分)
- `skills/synced/` は claude.ai から同期される組み込みスキルなので管理しない
- `git clean -fdx` は ignore 済みの履歴・セッションも消すので `common/claude` では使わない
