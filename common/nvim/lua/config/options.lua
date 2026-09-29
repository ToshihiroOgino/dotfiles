-- Options are automatically loaded before lazy.nvim startup
-- Default options that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/options.lua
-- Add any additional options here
if vim.fn.has("win32") == 1 then
  LazyVim.terminal.setup("pwsh")
end

vim.o.relativenumber = true
vim.o.fileformat = "unix"
vim.o.fileformats = "unix,dos"
vim.o.wrap = true
vim.o.tabstop = 2
vim.o.swapfile = false
vim.o.scrolloff = 10
vim.o.backup = false
vim.o.writebackup = false
