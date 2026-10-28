vim.pack.add({
	"https://github.com/mason-org/mason.nvim",
	"https://github.com/WhoIsSethDaniel/mason-tool-installer.nvim",
}, { confirm = false })

require("mason").setup()

require("mason-tool-installer").setup({
	ensure_installed = {
		-- go
		"gopls",
		-- lua
		"lua-language-server",
		"stylua",
		-- shell
		"bash-language-server",
		"shfmt",
		"shellcheck",
		-- toml
		"taplo",
		-- json, js, ts
		"typescript-language-server",
		"json-lsp",
		"biome",
		-- markdown, yaml
		"prettier",
	},
	start_delay = 3000,
})
