vim.g.mapleader = " "

-- vim と共通の設定は config/vim/vimrc に置く
local vimrc = vim.fn.fnamemodify(vim.fn.resolve(vim.fn.stdpath("config")), ":h") .. "/vim/vimrc"
if vim.fn.filereadable(vimrc) == 1 then
	vim.cmd.source(vim.fn.fnameescape(vimrc))
end

vim.o.fileformat = "unix"

-- Run :terminal and shell commands in pwsh on Windows.
if vim.fn.has("win32") == 1 then
	vim.o.shell = "pwsh"
	vim.o.shellcmdflag =
		"-NoLogo -NonInteractive -ExecutionPolicy RemoteSigned -Command [Console]::InputEncoding=[Console]::OutputEncoding=[System.Text.UTF8Encoding]::new();$PSStyle.OutputRendering='plaintext';Remove-Alias -Force -ErrorAction SilentlyContinue tee;"
	vim.o.shellredir = '2>&1 | %%{ "$_" } | Out-File %s; exit $LastExitCode'
	vim.o.shellpipe = '2>&1 | %%{ "$_" } | tee %s; exit $LastExitCode'
	vim.o.shellquote = ""
	vim.o.shellxquote = ""
end

-- Tree-sitter でフォールドする。foldcolumn をクリックで開閉できる
vim.o.foldmethod = "expr"
vim.o.foldexpr = "v:lua.vim.treesitter.foldexpr()"
vim.o.foldcolumn = "1"
vim.o.foldlevelstart = 99
