vim.keymap.set("n", "q", "<Nop>", { desc = "Disable macro recording" })
vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Leave insert mode (fires InsertLeave)" })
vim.keymap.set("n", "<Esc>", "<cmd>nohlsearch<CR>", { silent = true, desc = "Clear search highlight" })
vim.keymap.set("i", "<C-CR>", "<C-o>o", { desc = "Open line below" })
vim.keymap.set("i", "<C-S-CR>", "<C-o>O", { desc = "Open line above" })
vim.keymap.set({ "n", "i" }, "<C-s>", "<Cmd>write<CR><Esc>", { desc = "Save and return to normal mode" })

vim.keymap.set("n", "<leader>ff", function()
	Snacks.picker.files()
end, { desc = "Find files" })
vim.keymap.set("n", "<leader>/", function()
	Snacks.picker.grep()
end, { desc = "Grep" })
vim.keymap.set("n", "<leader>fb", function()
	Snacks.picker.buffers()
end, { desc = "Buffers" })
vim.keymap.set("n", "<leader>fr", function()
	Snacks.picker.recent({ filter = { cwd = true } })
end, { desc = "Recent files (cwd)" })
vim.keymap.set("n", "<leader>fR", function()
	Snacks.picker.recent()
end, { desc = "Recent files" })
vim.keymap.set("n", "<leader>e", "<Cmd>Oil<CR>", { desc = "Open parent directory" })
local function run(cmd, cwd)
	local res = vim.system(cmd, { cwd = cwd, text = true }):wait()
	if res.code ~= 0 then
		return nil
	end
	local out = vim.trim(res.stdout or "")
	return out ~= "" and out or nil
end

local function buffer_git_root(silent)
	local file = vim.api.nvim_buf_get_name(0)
	local root
	if vim.bo.buftype == "" and file ~= "" then
		local dir = vim.fn.fnamemodify(file, ":p:h")
		if vim.fn.isdirectory(dir) == 1 then
			root = run({ "git", "rev-parse", "--show-toplevel" }, dir)
		end
	end
	root = root or run({ "git", "rev-parse", "--show-toplevel" }, vim.fn.getcwd())
	if not root and not silent then
		vim.notify("Not a git repository", vim.log.levels.WARN)
	end
	return root
end

local function diffview(subcmd, args)
	local root = buffer_git_root()
	if root then
		vim.cmd(subcmd .. " -C" .. vim.fn.fnameescape(root) .. " " .. (args or ""))
	end
end

vim.keymap.set("n", "<leader>gg", function()
	Snacks.lazygit({ cwd = buffer_git_root(true) })
end, { desc = "Lazygit" })
vim.keymap.set("n", "<leader>gd", function()
	diffview("DiffviewOpen")
end, { desc = "Diffview: working tree" })
vim.keymap.set("n", "<leader>gm", function()
	local root = buffer_git_root()
	if not root then
		return
	end
	local base
	if vim.fn.executable("gh") == 1 then
		local pr_base = run({ "gh", "pr", "view", "--json", "baseRefName", "-q", ".baseRefName" }, root)
		if pr_base and run({ "git", "rev-parse", "--verify", "--quiet", "origin/" .. pr_base }, root) then
			base = "origin/" .. pr_base
		end
	end
	base = base or run({ "git", "rev-parse", "--abbrev-ref", "origin/HEAD" }, root) or "main"
	diffview("DiffviewOpen", base .. "...HEAD --imply-local")
end, { desc = "Diffview: branch vs base" })
vim.keymap.set("n", "<leader>gh", function()
	local file = vim.api.nvim_buf_get_name(0)
	local is_file = vim.bo.buftype == "" and file ~= "" and vim.fn.filereadable(file) == 1
	diffview("DiffviewFileHistory", is_file and vim.fn.fnameescape(vim.fn.fnamemodify(file, ":p")) or "")
end, { desc = "Diffview: file history" })
vim.keymap.set("n", "<leader>gH", function()
	diffview("DiffviewFileHistory")
end, { desc = "Diffview: repo history" })
vim.keymap.set("n", "<leader>gq", "<Cmd>DiffviewClose<CR>", { desc = "Diffview: close" })
vim.keymap.set({ "n", "x" }, "<leader>cf", function()
	require("conform").format()
end, { desc = "Format" })
vim.keymap.set({ "n", "x", "i" }, "<M-F>", function()
	require("conform").format()
end, { desc = "Format" })

vim.api.nvim_create_autocmd("LspAttach", {
	group = vim.api.nvim_create_augroup("lsp_keymaps", { clear = true }),
	callback = function(args)
		vim.keymap.set("n", "gd", vim.lsp.buf.definition, { buffer = args.buf, desc = "Definition" })
		for _, lhs in ipairs({ "grn", "<F2>", "<leader>rn" }) do
			vim.keymap.set("n", lhs, vim.lsp.buf.rename, { buffer = args.buf, desc = "Rename symbol" })
		end
	end,
})

-- treesitter-textobjects: select (operator-pending / visual)
local ts_select = {
	af = "@function.outer",
	["if"] = "@function.inner",
	ac = "@class.outer",
	ic = "@class.inner",
	aa = "@parameter.outer",
	ia = "@parameter.inner",
	al = "@loop.outer",
	il = "@loop.inner",
	ai = "@conditional.outer",
	ii = "@conditional.inner",
}
for lhs, query in pairs(ts_select) do
	vim.keymap.set({ "x", "o" }, lhs, function()
		require("nvim-treesitter-textobjects.select").select_textobject(query, "textobjects")
	end, { desc = query })
end

-- treesitter-textobjects: move
local ts_move = {
	["]f"] = { "goto_next_start", "@function.outer" },
	["[f"] = { "goto_previous_start", "@function.outer" },
	["]F"] = { "goto_next_end", "@function.outer" },
	["[F"] = { "goto_previous_end", "@function.outer" },
	["]a"] = { "goto_next_start", "@parameter.inner" },
	["[a"] = { "goto_previous_start", "@parameter.inner" },
	["]l"] = { "goto_next_start", "@loop.outer" },
	["[l"] = { "goto_previous_start", "@loop.outer" },
}
for lhs, spec in pairs(ts_move) do
	vim.keymap.set({ "n", "x", "o" }, lhs, function()
		require("nvim-treesitter-textobjects.move")[spec[1]](spec[2], "textobjects")
	end, { desc = spec[2] })
end

-- treesitter-textobjects: swap
vim.keymap.set("n", "<leader>sa", function()
	require("nvim-treesitter-textobjects.swap").swap_next("@parameter.inner")
end, { desc = "Swap parameter with next" })
vim.keymap.set("n", "<leader>sA", function()
	require("nvim-treesitter-textobjects.swap").swap_previous("@parameter.inner")
end, { desc = "Swap parameter with previous" })

-- treesitter-context: jump to the line that is pinned at the top
vim.keymap.set("n", "<leader>cx", function()
	require("treesitter-context").go_to_context(vim.v.count1)
end, { desc = "Go to context" })

-- buffers: close without breaking the window layout (snacks bufdelete)
vim.keymap.set("n", "<leader>bd", function()
	Snacks.bufdelete()
end, { desc = "Delete buffer" })
vim.keymap.set("n", "<leader>bo", function()
	Snacks.bufdelete.other()
end, { desc = "Delete other buffers" })
vim.keymap.set("n", "<leader>ba", function()
	Snacks.bufdelete.all()
end, { desc = "Delete all buffers" })
