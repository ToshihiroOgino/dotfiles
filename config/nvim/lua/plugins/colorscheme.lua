local active = "catppuccin" -- "catppuccin" | "gruvbox"

vim.pack.add({
  { src = "https://github.com/catppuccin/nvim", name = "catppuccin" },
  "https://github.com/ellisonleao/gruvbox.nvim",
}, { confirm = false })

require("catppuccin").setup({ flavour = "mocha" }) -- latte | frappe | macchiato | mocha
require("gruvbox").setup({ contrast = "hard" })

vim.cmd.colorscheme(active)
