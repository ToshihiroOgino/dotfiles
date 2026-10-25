# dotfiles

## Requirements

- make
- go (to build `imestat` and `pkgsync`)
- Windows Developer Mode (needed to create symlinks)

## Usage

```sh
make link            # link all configs for this OS
make imestat         # build imestat
make pkgsync         # build pkgsync
make systemd-enable  # WSL only: enable systemd units
```

On Windows, recipes need sh (Git Bash).

## Tools

- [imestat](tools/imestat/README.md): shows the IME mode in the tmux status bar.
- [pkgsync](tools/pkgsync/README.md): compares installed packages with packages/<mgr>.json.
