vim.pack.add({ "https://github.com/stevearc/oil.nvim" }, { confirm = false })

require("oil").setup({
  view_options = { show_hidden = true },
})
