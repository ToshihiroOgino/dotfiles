# pkgsync

Compares the installed packages with packages/<mgr>.json.

Looks for ~/dotfiles, or $PKGSYNC_ROOT if set.

## Build

```sh
make build
```

## Run

```sh
./pkgsync init    # create empty packages/<mgr>.json for all managers
./pkgsync         # show the difference
./pkgsync add     # add new packages to the list
./pkgsync sync    # add new packages and drop removed ones
./pkgsync help    # list the commands
./pkgsync check   # for shell startup: refresh packages/pending/ (once per 5 min)
```

## Platforms

- Linux (apt)
- Windows (scoop)
