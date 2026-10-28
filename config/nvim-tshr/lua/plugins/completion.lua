vim.g.completion_enabled = true

local function toggle()
	vim.g.completion_enabled = not vim.g.completion_enabled
	vim.notify("completion: " .. (vim.g.completion_enabled and "on" or "off"))
end

local function setup_native()
	vim.o.autocomplete = true
	vim.o.autocompletedelay = 5
	vim.o.complete = "o,.,w,b"
	vim.o.completeopt = "menu,menuone,popup,fuzzy,noselect"

	vim.keymap.set("i", "<Tab>", function()
		return vim.fn.pumvisible() == 1 and "<C-n>" or "<Tab>"
	end, { expr = true, desc = "Next completion or Tab" })
	vim.keymap.set("i", "<S-Tab>", function()
		return vim.fn.pumvisible() == 1 and "<C-p>" or "<S-Tab>"
	end, { expr = true, desc = "Previous completion or S-Tab" })
	vim.keymap.set("i", "<M-i>", function()
		toggle()
		vim.o.autocomplete = vim.g.completion_enabled
	end, { desc = "Toggle completion" })
end

local function setup_blink()
	require("blink.cmp").setup({
		enabled = function()
			return vim.g.completion_enabled
		end,
		keymap = {
			preset = "default",
			["<C-space>"] = {},
			["<Tab>"] = { "select_next", "fallback" },
			["<S-Tab>"] = { "select_prev", "fallback" },
			["<M-i>"] = {
				function(cmp)
					toggle()
					cmp.hide()
					return true
				end,
			},
		},
		completion = {
			list = { selection = { preselect = false, auto_insert = false } },
			documentation = { auto_show = true },
		},
	})
	vim.lsp.config("*", { capabilities = require("blink.cmp").get_lsp_capabilities() })
end

local ok, err = pcall(vim.pack.add, {
	{ src = "https://github.com/saghen/blink.cmp", version = vim.version.range("1.*") },
}, { confirm = false })
if ok then
	ok, err = pcall(setup_blink)
end

if not ok then
	vim.notify("blink.cmp is not available, using the native completion:\n" .. tostring(err), vim.log.levels.WARN)
	setup_native()
end
