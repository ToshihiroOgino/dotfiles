vim.pack.add({ "https://github.com/folke/snacks.nvim" }, { confirm = false })

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

-- lazygit writes its theme into the cache dir, which nvim does not create by itself
vim.fn.mkdir(vim.fn.stdpath("cache"), "p")

require("snacks").setup({
  image = { enabled = true },
  picker = {
    sources = {
      explorer = source(),
      files = source(),
      grep = source(),
    },
  },
})
