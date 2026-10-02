# Herdr

herdr は git の `include` に相当する設定ファイルの取り込みに対応していないので、
共通設定と OS 固有設定を分けて書き、マージした結果を `generated/` に出力している。
設定ディレクトリには `generated/` のファイルを symlink する。

## ファイル構成

| ファイル | 役割 |
| --- | --- |
| `config.base.toml` | OS 非依存の設定。基本はここを編集する |
| `windows.toml` / `linux.toml` / `macos.toml` | OS 固有の設定。シェルのパスなど |
| `generated/config.<os>.toml` | 生成物。直接編集しない |
| `gen-config.lua` | マージして生成物を書くスクリプト |

設定可能なキーの一覧は `herdr --default-config` で確認する。

`generated/` はコミットしてある。clone 直後に symlink が切れないようにするため。

## 生成

このディレクトリでもリポジトリルートでも実行できる (ルートの Makefile が委譲する)。

```sh
make            # 全OS分
make windows    # Windows 分だけ
make check      # 設定ディレクトリの symlink 経由で herdr のバリデータに通す
make clean      # 生成物を削除
```

`lua` が無い環境では `make LUA="nvim -l"` で代用できる。

編集するのは `config.base.toml` と `<os>.toml` だけ。生成物を書き換えても次の
`make` で消える。

生成後は `herdr server reload-config` (または prefix+shift+R) で起動中のサーバに
反映する。サーバは設定を起動時に読むだけなので、これを叩かないと生成しても変化が
ない。既存のペインは作り直されないので、`default_shell` の変更は次に作るペインから
効く。

マージ規則は `gen-config.lua` の冒頭コメントを参照。要点は「同じテーブル内の
同名キーはオーバーレイが勝つ」「`[[array.of.tables]]` はオーバーレイが定義したら
丸ごと置き換え」。

## symlink

- Windows: `$env:APPDATA\herdr\config.toml` -> `generated/config.windows.toml`
- Linux / macOS: `~/.config/herdr/config.toml` -> `generated/config.<os>.toml`

```powershell
# Windows (開発者モードか管理者権限が必要)
New-Item -ItemType SymbolicLink -Path "$env:APPDATA\herdr\config.toml" `
  -Target "$env:USERPROFILE\dotfiles\common\herdr\generated\config.windows.toml"
```

```sh
# Linux
ln -sfn ~/dotfiles/common/herdr/generated/config.linux.toml ~/.config/herdr/config.toml
```

## シェルの決定

`default_shell` を空にしておくと herdr が自前で選ぶ。Windows は PATH から
`pwsh` -> `powershell` の順に探し、Unix は `$SHELL` -> `/bin/sh` の順。
Windows では `$SHELL` を見ないので、Git Bash から起動しても pwsh になる。

`<os>.toml` で明示しているのはバージョンを固定したいから。Windows は
PowerShell 5.1 ではなく 7 を確実に引くため絶対パスで、Unix は PATH 解決される
ので `zsh` と名前だけ書いている。
