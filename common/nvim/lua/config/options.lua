-- Options are automatically loaded before lazy.nvim startup
-- Default options that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/options.lua
-- Add any additional options here
if vim.fn.has("win32") == 1 then
  LazyVim.terminal.setup("pwsh")
end

-- vim と共通の設定は common/vim/vimrc に置く
local vimrc = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.stdpath("config")), ":h") .. "/vim/vimrc"
if vim.fn.filereadable(vimrc) == 1 then
  vim.cmd.source(vim.fn.fnameescape(vimrc))
end

-- nvim 固有 (バッファローカルのため vimrc では効果が薄い)
vim.o.fileformat = "unix"
