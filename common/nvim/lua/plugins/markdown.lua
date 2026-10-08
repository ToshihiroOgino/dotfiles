return {
  {
    "mfussenegger/nvim-lint",
    opts = function()
      local lint = require("lint")
      local linter = lint.linters["markdownlint-cli2"]
      local config = vim.fn.stdpath("config") .. "/markdownlint-cli2.yaml"
      linter.args = vim.list_extend({ "--config", config }, linter.args or {})
    end,
  },
}
