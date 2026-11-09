vim.pack.add({ "https://github.com/sindrets/diffview.nvim" }, { confirm = false })

require("diffview").setup({
	enhanced_diff_hl = true,
	use_icons = false,
})
