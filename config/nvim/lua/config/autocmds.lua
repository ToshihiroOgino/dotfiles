require("config.ime")

vim.api.nvim_create_autocmd("BufReadPost", {
	group = vim.api.nvim_create_augroup("readonly_binary", { clear = true }),
	callback = function(args)
		local name = vim.api.nvim_buf_get_name(args.buf)
		local is_binary = vim.bo[args.buf].binary or name:lower():match("%.pdf$") ~= nil
		if not is_binary then
			local f = io.open(name, "rb")
			if f then
				is_binary = (f:read(8192) or ""):find("\0", 1, true) ~= nil
				f:close()
			end
		end
		if not is_binary then
			return
		end
		vim.bo[args.buf].readonly = true
		vim.bo[args.buf].modifiable = false
		vim.bo[args.buf].modified = false
	end,
})

vim.api.nvim_create_autocmd("VimResized", {
	group = vim.api.nvim_create_augroup("equalize_on_resize", { clear = true }),
	callback = function()
		local tab = vim.api.nvim_get_current_tabpage()
		vim.cmd("tabdo wincmd =")
		vim.api.nvim_set_current_tabpage(tab)
	end,
})
