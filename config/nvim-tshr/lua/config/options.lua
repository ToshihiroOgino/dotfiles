vim.g.mapleader = " "

-- vim と共通の設定は config/vim/vimrc に置く
local vimrc = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.stdpath("config")), ":h") .. "/vim/vimrc"
if vim.fn.filereadable(vimrc) == 1 then
  vim.cmd.source(vim.fn.fnameescape(vimrc))
end

vim.o.fileformat = "unix"

-- Tree-sitter でフォールドする。foldcolumn をクリックで開閉できる
vim.o.foldmethod = "expr"
vim.o.foldexpr = "v:lua.vim.treesitter.foldexpr()"
vim.o.foldcolumn = "1"
vim.o.foldlevelstart = 99
