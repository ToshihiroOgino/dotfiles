vim.pack.add({
  { src = "https://github.com/nvim-treesitter/nvim-treesitter", version = "main" },
  { src = "https://github.com/nvim-treesitter/nvim-treesitter-textobjects", version = "main" },
  "https://github.com/nvim-treesitter/nvim-treesitter-context",
}, { confirm = false })

-- Parsers for the languages used in dotfiles. Installs are async and skip what is already there.
-- c, lua, vim, vimdoc, query and markdown ship with nvim itself.
require("nvim-treesitter").install({
  "go",
  "gomod",
  "bash",
  "toml",
  "json",
  "yaml",
  "powershell",
  "javascript",
  "typescript",
  "tsx",
  "make",
})

-- Highlighting is provided by nvim; start it per buffer (ignored when no parser exists).
vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("treesitter_start", { clear = true }),
  callback = function(args)
    pcall(vim.treesitter.start, args.buf)
  end,
})

require("nvim-treesitter-textobjects").setup({
  select = { lookahead = true },
  move = { set_jumps = true },
})

require("treesitter-context").setup({ max_lines = 3 })
