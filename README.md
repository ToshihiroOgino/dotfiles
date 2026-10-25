# dotfiles

## Requirements

- make
- Git for Windows
- go
- Windows Developer Mode

## Usage

```sh
make all             # everything below
make link            # link all configs for this OS
make imestat         # build imestat
make pkgsync         # build pkgsync
make systemd-enable  # WSL only: enable systemd units
```

## Tools

- [imestat](tools/imestat/README.md): shows the IME mode in the tmux status bar.
- [pkgsync](tools/pkgsync/README.md): compares installed packages with packages/<manager>.json.
