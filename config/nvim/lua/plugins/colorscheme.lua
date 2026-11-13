local active = "catppuccin" -- "catppuccin" | "gruvbox" | "github_dark_default"

vim.pack.add({
  { src = "https://github.com/catppuccin/nvim", name = "catppuccin" },
  "https://github.com/ellisonleao/gruvbox.nvim",
  "https://github.com/projekt0n/github-nvim-theme",
}, { confirm = false })

require("catppuccin").setup({ flavour = "macchiato" }) -- latte | frappe | macchiato | mocha
require("gruvbox").setup({ contrast = "hard" })

vim.cmd.colorscheme(active)
