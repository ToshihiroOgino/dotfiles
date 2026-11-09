require("config.ime")

vim.api.nvim_create_autocmd("VimResized", {
	group = vim.api.nvim_create_augroup("equalize_on_resize", { clear = true }),
	callback = function()
		local tab = vim.api.nvim_get_current_tabpage()
		vim.cmd("tabdo wincmd =")
		vim.api.nvim_set_current_tabpage(tab)
	end,
})
