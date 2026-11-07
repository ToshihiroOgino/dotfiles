vim.pack.add({ "https://github.com/stevearc/conform.nvim" }, { confirm = false })

require("conform").setup({
	default_format_opts = { lsp_format = "fallback" },
	formatters_by_ft = {
		lua = { "stylua" },
		sh = { "shfmt" },
		python = { "ruff_organize_imports", "ruff_format" },
		markdown = { "prettier" },
		yaml = { "prettier" },
		json = { "biome" },
		jsonc = { "biome" },
		javascript = { "biome" },
		typescript = { "biome" },
		["*"] = { "trim_whitespace" },
		["_"] = { lsp_format = "first" },
	},
})
