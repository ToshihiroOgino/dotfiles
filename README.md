# dotfiles

## Requirements

- make
- go (to build `imestat`)
- Windows Developer Mode (needed to create symlinks)

## Usage

```sh
make link            # link all configs for this OS
make imestat         # build imestat
make systemd-enable  # WSL only: enable systemd units
```

On Windows, recipes need sh (Git Bash).

## Tools

- [imestat](tools/imestat/README.md): shows the IME mode in the tmux status bar.
