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
make ime        # build ime
make pkgsync         # build pkgsync
make systemd-enable  # WSL only: enable systemd units
```

## Tools

- [ime](tools/ime/README.md): reads and switches the IME mode (tmux status bar, IME off during prefix).
- [pkgsync](tools/pkgsync/README.md): compares installed packages with packages/<manager>.json.
