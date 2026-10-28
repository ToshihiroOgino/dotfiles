vim.pack.add({ "https://github.com/stevearc/oil.nvim" }, { confirm = false })

require("oil").setup({
	-- oil has its own trash for each OS (freedesktop trash, ~/.Trash, Recycle Bin), no trash-cli needed.
	delete_to_trash = true,
	view_options = { show_hidden = true },
})
