# imestat

Reports the IME input mode of the foreground window for a tmux status line.

| Mode              | Usage                                                       |
| ----------------- | ----------------------------------------------------------- |
| One-shot          | `imestat` prints the mode once                              |
| Watch (resident)  | `imestat --watch` pushes the mode into tmux option `@ime`   |

| Output | Meaning                                   |
| ------ | ----------------------------------------- |
| `あ`   | IME on, native (kana) input mode          |
| `A`    | IME off, or on in alphanumeric mode       |
| `?`    | State could not be read (e.g. no window)  |

On unsupported platforms nothing is written to stdout; `imestat: ... not supported on <os>`
goes to stderr and the exit code is 1.

## Platform Support

| OS      | Status                                  |
| ------- | --------------------------------------- |
| Windows | Supported (psmux)                       |
| Linux   | Stub (builds, exits 1)                  |
| macOS   | Stub (builds, exits 1)                  |

Platform code lives in `*_windows.go` / `*_other.go` files selected by build
constraints. To add an OS, add `ime_<os>.go` / `watch_<os>.go` and narrow the
constraint in the `*_other.go` files.

## Build / Install

Requires Go. Recipes only call `go`, so any `make` works (cmd, pwsh, or sh).

```sh
make build    # ./imestat(.exe)
make install  # go install -> $(go env GOBIN) or ~/go/bin (must be on PATH)
make clean    # remove build output and the installed binary
```

Windows locks a running executable, so stop the watchers before reinstalling
(`tmux kill-server`, or `Stop-Process -Name imestat`).

## tmux

Configured in [common/tmux/tmux.conf](../../common/tmux/tmux.conf):

```tmux
if-shell 'imestat' "run-shell -b 'imestat --watch'"
set -g status-right-length 50
set -g status-right '...#{@ime} "#{=21:pane_title}" %H:%M %d-%b-%y'
```

`if-shell` fails quietly where imestat is missing or unsupported, so the same
config works everywhere; `#{@ime}` is just empty there.

## How It Works (Windows)

### Reading the IME state

1. `GetForegroundWindow` - the window currently in front (e.g. Alacritty).
2. `ImmGetDefaultIMEWnd` - the hidden IME window attached to it.
3. `SendMessageTimeoutW(WM_IME_CONTROL, IMC_GETOPENSTATUS)` - IME on/off.
4. `SendMessageTimeoutW(WM_IME_CONTROL, IMC_GETCONVERSIONMODE)` - conversion mode.

`あ` requires both "open" and `IME_CMODE_NATIVE`, because a modern MS-IME can
report "open" while in half-width alphanumeric mode. Messages use a 50 ms timeout
with `SMTO_ABORTIFHUNG`, so a hung foreground app yields `?` instead of blocking.
Only `syscall` is used; no cgo or C compiler is needed.

### Watch mode

- Polls the state in-process every 100 ms (no process spawn) and runs
  `tmux set-option -t <session> -g @ime <mode>` only when the mode changes.
  Setting an option repaints the status line immediately, so `status-interval`
  plays no part.
- psmux runs one server per session and `set -g` reaches only one of them, so
  every session is targeted by name. Every 2 s it also looks for new sessions
  and sends the current mode to those only.
- Each psmux server runs the config, so each starts a watcher. A named mutex
  (`Local\imestat-watch`) lets one work while the others wait to take over.
  A per-server mutex drops duplicates (psmux re-runs the config when a warm
  server is claimed).
- psmux starts run-shell commands via `pwsh -Command` and leaves them running
  on `kill-server`. The watcher finds its server by walking the process tree,
  re-launches itself detached with `--server-pid`, and exits, so the ~75 MB
  pwsh wrapper goes away. The detached watcher (~6 MB) exits when that server
  process does.
- Without a psmux server among its ancestors, `--watch` exits with an error
  instead of running forever. Pass `--server-pid <pid>` to run it by hand.
- `-L <socket>` targets another tmux socket (used for testing).

## Limitations

- Up to 100 ms delay in watch mode (polling interval).
- Only the default socket is updated. run-shell children get no `TMUX` and no
  format expansion, so the watcher cannot tell which socket its server uses.
  A server started with `-L <name>` needs `imestat --watch -L <name>`.
- It reports the foreground window, so while another app is focused the
  indicator shows that app's IME state.
