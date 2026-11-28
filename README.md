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
make tools           # build the tools below and link them into ~/.local/bin
make systemd-enable  # WSL only: enable systemd units
make external        # fetch externally hosted tools and put them on PATH
```

## Tools

- [pkgsync](tools/pkgsync/README.md): compares installed packages with packages/<manager>.json.
- claude-tmux-state: shows Claude Code's state in the tmux / psmux window list.
