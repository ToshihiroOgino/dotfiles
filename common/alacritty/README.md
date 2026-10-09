# Alacritty

## Config File Location

| OS      | Path                                 |
| ------- | ------------------------------------ |
| Windows | `%APPDATA%\alacritty\alacritty.toml` |

Instead of symlinking, import this file from the config above.
Machine specific settings (e.g. `shell`) stay in the local config and
override the imported values.

```toml
[general]
import = ["~/dotfiles/common/alacritty/alacritty.toml"]

[terminal]
shell = "pwsh.exe"
```

`~/` is expanded by Alacritty itself, so the same line works on any machine
that clones this repo into `~/dotfiles`.

## macOS

macOS では `~/.config/alacritty/alacritty.toml` から、共通設定に加えて
macOS 専用設定(Option+←/→ での単語移動)を import する。

```toml
[general]
import = [
  "~/dotfiles/common/alacritty/alacritty.toml",
  "~/dotfiles/common/alacritty/mac.toml",
]
```
