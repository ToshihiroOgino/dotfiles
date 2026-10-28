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

On Linux and macOS, `ime prefix` hands the work to the running `ime watch` over a Unix socket (`$XDG_RUNTIME_DIR/ime-prefix[-socket].sock`, or a private `ime-<uid>` directory under the temporary directory). The watcher keeps one `tmux -C` connection and asks it for the client's key table while the prefix is active, so no tmux process starts per poll. Without a watcher or a control connection (and on Windows), `ime prefix` polls `tmux list-clients` itself.

## Layout

- `internal/imectl`: reads and switches the IME mode, one file per platform.
- `internal/tmuxctl`: runs tmux commands, per call (`Client`) or over one control-mode connection (`Pool`), and parses `list-clients`.
- `internal/watch`: keeps tmux `@ime` updated; per-platform change detection.
- `internal/prefix`: the prefix session, and the socket that hands it to the watcher.
- the main package: the command line.

## Platforms

- Windows
- macOS
- Linux (fcitx5; needs fcitx5-remote and dbus-monitor)
