-- vim と共通の設定は config/vim/vimrc に置く
local vimrc = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.stdpath("config")), ":h") .. "/vim/vimrc"
if vim.fn.filereadable(vimrc) == 1 then
  vim.cmd.source(vim.fn.fnameescape(vimrc))
end

-- nvim 固有 (バッファローカルのため vimrc では効果が薄い)
vim.o.fileformat = "unix"
