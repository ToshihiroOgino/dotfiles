vim.pack.add({ "https://github.com/MeanderingProgrammer/render-markdown.nvim" }, { confirm = false })

require("render-markdown").setup({
	code = { border = "thin" },
	heading = {
		-- Keep the '#' markers as they are instead of replacing them with icons.
		icons = function(context)
			return string.rep("#", context.level) .. " "
		end,
		position = "inline",
		sign = false,
	},
})
