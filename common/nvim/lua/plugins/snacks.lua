-- Show dotfiles everywhere; hide only what is listed in `exclude` (blacklist).
local exclude = {
  ".git/", -- add more entries here, e.g. ".venv/"
}

return {
  "folke/snacks.nvim",
  opts = {
    picker = {
      sources = {
        explorer = { hidden = true, exclude = exclude },
        files = { hidden = true, exclude = exclude },
        grep = { hidden = true, exclude = exclude },
      },
    },
  },
}
