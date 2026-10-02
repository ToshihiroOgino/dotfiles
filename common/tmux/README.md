# tmux

## Config File Location

Symlink `tmux.conf` to `~/.tmux.conf` on every OS.
psmux (Windows native tmux) also reads `~/.tmux.conf`.

```sh
# Linux / macOS / WSL
ln -s ~/dotfiles/common/tmux/tmux.conf ~/.tmux.conf
```

```powershell
# Windows (psmux)
New-Item -ItemType SymbolicLink -Path ~/.tmux.conf -Target ~/dotfiles/common/tmux/tmux.conf
```

## IME Indicator

`status-right` shows the IME mode (`あ` / `A`) from `@ime`, which a resident
[imestat](../../tools/imestat) watcher updates only when the mode changes.
Install it onto `PATH` with `make install` in that directory.
It currently works on Windows only; elsewhere the indicator is simply empty.
