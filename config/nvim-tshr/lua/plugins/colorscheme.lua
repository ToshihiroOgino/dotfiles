vim.pack.add({ "https://github.com/ellisonleao/gruvbox.nvim" }, { confirm = false })

require("gruvbox").setup({ contrast = "hard" })
vim.cmd.colorscheme("gruvbox")
