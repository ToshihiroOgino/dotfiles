# nvim-tshr

Neovim config written from scratch, run with `NVIM_APPNAME=nvim-tshr`. Plugins are managed by the built-in `vim.pack`; `nvim-pack-lock.json` pins their revisions.

## Requirements

- Neovim 0.12 (`make nvim`)
- git, curl, tar, a C compiler
- tree-sitter CLI 0.26.1 or later (`make tree-sitter`)
- node and npm (mason installs some language servers with npm)
- lazygit, ripgrep, fd
- ime (`make tools`), optional
- trash-cli, optional: oil deletes to the trash when it is installed

## Usage

```sh
make link-nvim-tshr   # link this directory to ~/.config/nvim-tshr
v                     # alias for NVIM_APPNAME=nvim-tshr nvim (config/zsh/zshrc)
```

Language servers and tools listed in `lua/plugins/mason.lua` are installed in the background on startup.

## Keybinds

The leader key is Space.

### Find

| Key          | Action       |
| ------------ | ------------ |
| `<leader>ff` | files        |
| `<leader>/`  | grep         |
| `<leader>fb` | buffers      |
| `<leader>fr` | recent files |

### Files and git

| Key          | Action                                                        |
| ------------ | ------------------------------------------------------------- |
| `<leader>e`  | oil: open the parent directory (`-` goes up, `g?` shows help) |
| `<leader>gg` | lazygit                                                       |

### Buffers

| Key          | Action                                |
| ------------ | ------------------------------------- |
| `<leader>bd` | delete the buffer                     |
| `<leader>bo` | delete the other buffers              |
| `<leader>ba` | delete all buffers                    |
| `[b` `]b`    | previous / next buffer (nvim default) |

### Code

| Key                                    | Action                                                                               |
| -------------------------------------- | ------------------------------------------------------------------------------------ |
| `gd`                                   | definition (LSP buffers)                                                             |
| `<leader>cf`                           | format (normal, visual)                                                              |
| `<leader>cx`                           | jump to the context line pinned at the top                                           |
| `K` `grr` `gri` `grt` `grn` `gra` `gO` | hover, references, implementation, type, rename, code action, symbols (nvim default) |

### Completion (insert mode)

| Key               | Action                 |
| ----------------- | ---------------------- |
| `<Tab>` `<S-Tab>` | next / previous item   |
| `<C-y>`           | accept                 |
| `<C-e>`           | hide                   |
| `<M-i>`           | toggle auto completion |
| `<C-c>`           | same as `<Esc>`        |

### Text objects (tree-sitter)

| Key                       | Action                                          |
| ------------------------- | ----------------------------------------------- |
| `af` `if`                 | function (outer / inner)                        |
| `ac` `ic`                 | class or type                                   |
| `aa` `ia`                 | parameter                                       |
| `al` `il`                 | loop                                            |
| `ai` `ii`                 | conditional                                     |
| `]f` `[f` `]F` `[F`       | next / previous function start / end            |
| `]a` `[a`                 | next / previous parameter                       |
| `]l` `[l`                 | next / previous loop                            |
| `<leader>sa` `<leader>sA` | swap the parameter with the next / previous one |

### Folds (nvim default)

| Key       | Action                           |
| --------- | -------------------------------- |
| `za`      | toggle the fold under the cursor |
| `zc` `zo` | close / open                     |
| `zM` `zR` | close / open all                 |
