vim.pack.add({ "https://github.com/stevearc/oil.nvim" }, { confirm = false })

local patterns = vim.tbl_map(vim.glob.to_lpeg, require("config.exclude"))

require("oil").setup({
	-- oil has its own trash for each OS (freedesktop trash, ~/.Trash, Recycle Bin), no trash-cli needed.
	delete_to_trash = true,
	view_options = {
		show_hidden = true,
		is_always_hidden = function(name)
			for _, pattern in ipairs(patterns) do
				if pattern:match(name) then
					return true
				end
			end
			return false
		end,
	},
})
