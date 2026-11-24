vim.pack.add({ "https://github.com/neovim/nvim-lspconfig" }, { confirm = false })

-- Teach lua_ls about nvim's runtime so `vim` is not an undefined global.
vim.lsp.config("lua_ls", {
	settings = {
		Lua = {
			runtime = { version = "LuaJIT" },
			workspace = { library = { vim.env.VIMRUNTIME }, checkThirdParty = false },
		},
	},
})

vim.lsp.config("typos_lsp", {
	init_options = {
		config = vim.fn.stdpath("config") .. "/typos.toml",
		diagnosticSeverity = "Hint",
	},
})

vim.diagnostic.config({
	virtual_text = false,
	virtual_lines = { current_line = true },
	severity_sort = true,
})

-- lspconfig names. The mason package names live in mason.lua.
vim.lsp.enable({
	"gopls",
	"clangd",
	"lua_ls",
	"bashls",
	"taplo",
	"jsonls",
	"biome", -- json, js, ts: lint and format
	"ts_ls", -- js, ts: definitions, references, completion
	"typos_lsp",
})
