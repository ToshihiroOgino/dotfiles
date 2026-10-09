-- Show dotfiles and gitignored files (e.g. logs/) everywhere;
-- hide only what is listed in `exclude` (blacklist).
local exclude = {
  ".git/",
  "node_modules/",
  ".venv/",
  "__pycache__/",
  ".next/",
  ".DS_Store",
}

local function source()
  return { hidden = true, ignored = true, exclude = exclude }
end

return {
  "folke/snacks.nvim",
  opts = {
    picker = {
      sources = {
        explorer = source(),
        files = source(),
        grep = source(),
      },
    },
  },
}
