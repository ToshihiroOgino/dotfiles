# imestat

Shows the IME mode in the tmux status bar.

Already configured. Start tmux and it shows up in the status bar.

## Build

```sh
make build
```

## Run

```sh
./imestat          # print the mode once
./imestat --watch  # keep tmux @ime updated (tmux.conf runs this)
```

## Platforms

- Windows
- macOS
- Linux (fcitx5; needs fcitx5-remote and dbus-monitor)
