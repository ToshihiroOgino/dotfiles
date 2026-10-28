vim.pack.add({ "https://github.com/stevearc/conform.nvim" }, { confirm = false })

-- Formatter names are conform's; the mason packages are listed in mason.lua.
-- Filetypes without an entry (go, toml) fall back to their LSP via
-- require("conform").format({ lsp_format = "fallback" }).
require("conform").setup({
  formatters_by_ft = {
    lua = { "stylua" },
    sh = { "shfmt" },
    markdown = { "prettier" },
    yaml = { "prettier" },
    json = { "biome" },
    jsonc = { "biome" },
    javascript = { "biome" },
    typescript = { "biome" },
  },
})
