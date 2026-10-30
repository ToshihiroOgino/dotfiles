vim.pack.add({ "https://github.com/folke/snacks.nvim" }, { confirm = false })

-- Show dotfiles and gitignored files (e.g. logs/) everywhere;
-- hide only what is listed in `exclude` (blacklist).
local exclude = require("config.exclude")

local function source()
	return { hidden = true, ignored = true, exclude = exclude }
end

-- lazygit writes its theme into the cache dir, which nvim does not create by itself
vim.fn.mkdir(vim.fn.stdpath("cache"), "p")

require("snacks").setup({
	image = {
		enabled = true,
		convert = {
			magick = {
				-- Default has "-trim", which crops PDF pages down to the text area.
				pdf = { "-density", 192, "{src}[{page}]", "-background", "white", "-alpha", "remove" },
			},
		},
	},
	picker = {
		sources = {
			explorer = source(),
			files = source(),
			grep = source(),
		},
	},
})
