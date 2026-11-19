vim.o.sessionoptions = "buffers,curdir,tabpages,winsize,winpos,folds"

local dir = vim.fn.stdpath("state") .. "/sessions"
local save_enabled = false

local function session_file()
	return dir .. "/" .. vim.fn.getcwd():gsub("[/\\:]", "%%") .. ".vim"
end

local function is_special(buf)
	return vim.bo[buf].buftype ~= "" or vim.api.nvim_buf_get_name(buf):match("^%w+://") ~= nil
end

local function has_file_buffer()
	for _, buf in ipairs(vim.api.nvim_list_bufs()) do
		if vim.bo[buf].buflisted and not is_special(buf) and vim.api.nvim_buf_get_name(buf) ~= "" then
			return true
		end
	end
	return false
end

local function save()
	if not save_enabled or not has_file_buffer() then
		return
	end
	for _, buf in ipairs(vim.api.nvim_list_bufs()) do
		if is_special(buf) then
			pcall(vim.api.nvim_buf_delete, buf, { force = true })
		end
	end
	vim.fn.mkdir(dir, "p")
	vim.cmd("mksession! " .. vim.fn.fnameescape(session_file()))
end

local function load(file)
	if vim.fn.filereadable(file) == 0 then
		vim.notify("no session", vim.log.levels.WARN)
		return
	end
	vim.cmd("silent! source " .. vim.fn.fnameescape(file))
	vim.cmd("wincmd =")
end

local function last_session()
	local files = vim.fn.glob(dir .. "/*.vim", false, true)
	table.sort(files, function(a, b)
		return vim.fn.getftime(a) > vim.fn.getftime(b)
	end)
	return files[1]
end

vim.api.nvim_create_autocmd("VimEnter", {
	group = vim.api.nvim_create_augroup("session", { clear = true }),
	nested = true,
	callback = function()
		save_enabled = vim.fn.argc() == 0
	end,
})

vim.api.nvim_create_autocmd("VimLeavePre", {
	group = "session",
	callback = save,
})

vim.keymap.set("n", "<leader>qs", function()
	load(session_file())
end, { desc = "Restore session (cwd)" })
vim.keymap.set("n", "<leader>ql", function()
	local file = last_session()
	if file then
		load(file)
	end
end, { desc = "Restore last session" })
vim.keymap.set("n", "<leader>qd", function()
	save_enabled = false
	vim.notify("session saving disabled")
end, { desc = "Don't save session" })
