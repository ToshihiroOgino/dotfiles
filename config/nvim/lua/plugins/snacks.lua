vim.pack.add({ "https://github.com/folke/snacks.nvim" }, { confirm = false })

local exclude = require("config.exclude")

local function source()
	return { hidden = true, ignored = false, exclude = exclude }
end

-- Run pdflatex via lualatex (bin/pdflatex shim) and load luatexja for Japanese in math.
vim.env.PATH = vim.fn.stdpath("config") .. "/bin:" .. vim.env.PATH

-- lazygit writes its theme into the cache dir, which nvim does not create by itself
vim.fn.mkdir(vim.fn.stdpath("cache"), "p")

require("snacks").setup({
	image = {
		enabled = true,
		math = {
			enabled = true,
			latex = {
				packages = { "amsmath", "amssymb", "amsfonts", "amscd", "mathtools", "luatexja" },
			},
		},
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
