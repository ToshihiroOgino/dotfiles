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
