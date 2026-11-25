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

vim.o.autoread = true

local checktime_group = vim.api.nvim_create_augroup("auto_checktime", { clear = true })

vim.api.nvim_create_autocmd({ "FocusGained", "BufEnter", "CursorHold", "TermLeave" }, {
	group = checktime_group,
	callback = function()
		if vim.fn.mode() ~= "c" then
			vim.cmd("silent! checktime")
		end
	end,
})

local checktime_timer = assert(vim.uv.new_timer())
checktime_timer:start(
	1000,
	1000,
	vim.schedule_wrap(function()
		if vim.fn.mode() ~= "c" then
			vim.cmd("silent! checktime")
		end
	end)
)

-- remove old undofiles
local undodir = vim.fn.stdpath("state") .. "/undo"
local undo_expire = os.time() - 30 * 24 * 60 * 60
vim.uv.fs_scandir(undodir, function(_, entries)
	if not entries then
		return
	end
	for name, type in vim.uv.fs_scandir_next, entries do
		if type == "file" then
			local path = vim.fs.joinpath(undodir, name)
			vim.uv.fs_stat(path, function(_, stat)
				if stat and stat.mtime.sec < undo_expire then
					vim.uv.fs_unlink(path, function() end)
				end
			end)
		end
	end
end)
