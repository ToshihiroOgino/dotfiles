-- Show dotfiles and gitignored files (e.g. logs/) everywhere;
-- hide only what is listed in `exclude` (blacklist).
local exclude = require("config.exclude")

local function source()
  return { hidden = true, ignored = true, exclude = exclude }
end

return {
  "folke/snacks.nvim",
  opts = {
    image = { enabled = true },
    picker = {
      sources = {
        explorer = source(),
        files = source(),
        grep = source(),
      },
    },
  },
}
