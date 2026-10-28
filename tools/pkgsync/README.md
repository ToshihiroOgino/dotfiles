# pkgsync

Compares the installed packages with packages/<mgr>.json.

Looks for ~/dotfiles, or $PKGSYNC_ROOT if set.

## Build

```sh
make build
```

## Run

```sh
./pkgsync init             # create empty packages/<mgr>.json for all managers
./pkgsync                  # show the difference and refresh packages/pending/
./pkgsync diff --throttle  # same, but skipped if the last successful run is under 5 min old (shell startup)
./pkgsync apply            # add new packages to the list and drop removed ones
./pkgsync help             # list the commands
```

`diff --throttle` is what the shell startup scripts run in the background,
with the output redirected.

## Files

- `packages/<mgr>.json`: the tracked list.
- `packages/pending/`: written by `diff` (`apply` clears it); `summary` is shown in the prompt.
  `.lock` and `.stamp` (time of the last successful `diff`) are also kept here.
- `log/pkgsync/<UTC time>.log`: written when a run fails. The next `pkgsync diff`
  shows the newest one until a run succeeds.

## Platforms

- Linux (apt)
- Windows (scoop)
- macOS (Homebrew: formulae from `brew leaves`, casks as `cask:<name>`, taps as buckets)
