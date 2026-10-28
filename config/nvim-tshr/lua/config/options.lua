vim.g.mapleader = " "

-- vim と共通の設定は config/vim/vimrc に置く
local vimrc = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.stdpath("config")), ":h") .. "/vim/vimrc"
if vim.fn.filereadable(vimrc) == 1 then
  vim.cmd.source(vim.fn.fnameescape(vimrc))
end

vim.o.fileformat = "unix"

vim.o.autocomplete = true
vim.o.autocompletedelay = 5
vim.o.complete = "o,.,w,b"
vim.o.completeopt = "menu,menuone,popup,fuzzy,noselect"
