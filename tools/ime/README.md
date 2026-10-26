# ime

Reads and switches the IME mode. Shows it in the tmux status bar, and switches to alphanumeric input while the tmux prefix (Ctrl+b) is active and restores it afterwards.

Already configured. Start tmux and it shows up in the status bar.

## Build

```sh
make build
```

## Run

```sh
./ime            # print the mode once (same as `get`)
./ime set en     # alphanumeric input
./ime set ja     # native (Japanese) input
./ime prefix CLIENT  # alphanumeric until the tmux client leaves the prefix table, then restore (tmux.conf runs this)
./ime watch      # keep tmux @ime updated (tmux.conf runs this)
```

## Platforms

- Windows
- macOS
- Linux (fcitx5; needs fcitx5-remote and dbus-monitor)
