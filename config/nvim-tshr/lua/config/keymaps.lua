vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Leave insert mode (fires InsertLeave)" })

vim.keymap.set("n", "<leader>ff", function() Snacks.picker.files() end, { desc = "Find files" })
vim.keymap.set("n", "<leader>/", function() Snacks.picker.grep() end, { desc = "Grep" })
vim.keymap.set("n", "<leader>fb", function() Snacks.picker.buffers() end, { desc = "Buffers" })
vim.keymap.set("n", "<leader>fr", function() Snacks.picker.recent() end, { desc = "Recent files" })
vim.keymap.set("n", "<leader>e", "<Cmd>Oil<CR>", { desc = "Open parent directory" })
vim.keymap.set("n", "<leader>gg", function() Snacks.lazygit() end, { desc = "Lazygit" })
vim.keymap.set({ "n", "x" }, "<leader>cf", function()
  require("conform").format({ lsp_format = "fallback" })
end, { desc = "Format" })

vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("lsp_keymaps", { clear = true }),
  callback = function(args)
    vim.keymap.set("n", "gd", vim.lsp.buf.definition, { buffer = args.buf, desc = "Definition" })
  end,
})

-- treesitter-textobjects: select (operator-pending / visual)
local ts_select = {
  af = "@function.outer",
  ["if"] = "@function.inner",
  ac = "@class.outer",
  ic = "@class.inner",
  aa = "@parameter.outer",
  ia = "@parameter.inner",
  al = "@loop.outer",
  il = "@loop.inner",
  ai = "@conditional.outer",
  ii = "@conditional.inner",
}
for lhs, query in pairs(ts_select) do
  vim.keymap.set({ "x", "o" }, lhs, function()
    require("nvim-treesitter-textobjects.select").select_textobject(query, "textobjects")
  end, { desc = query })
end

-- treesitter-textobjects: move
local ts_move = {
  ["]f"] = { "goto_next_start", "@function.outer" },
  ["[f"] = { "goto_previous_start", "@function.outer" },
  ["]F"] = { "goto_next_end", "@function.outer" },
  ["[F"] = { "goto_previous_end", "@function.outer" },
  ["]a"] = { "goto_next_start", "@parameter.inner" },
  ["[a"] = { "goto_previous_start", "@parameter.inner" },
  ["]l"] = { "goto_next_start", "@loop.outer" },
  ["[l"] = { "goto_previous_start", "@loop.outer" },
}
for lhs, spec in pairs(ts_move) do
  vim.keymap.set({ "n", "x", "o" }, lhs, function()
    require("nvim-treesitter-textobjects.move")[spec[1]](spec[2], "textobjects")
  end, { desc = spec[2] })
end

-- treesitter-textobjects: swap
vim.keymap.set("n", "<leader>sa", function()
  require("nvim-treesitter-textobjects.swap").swap_next("@parameter.inner")
end, { desc = "Swap parameter with next" })
vim.keymap.set("n", "<leader>sA", function()
  require("nvim-treesitter-textobjects.swap").swap_previous("@parameter.inner")
end, { desc = "Swap parameter with previous" })

-- treesitter-context: jump to the line that is pinned at the top
vim.keymap.set("n", "<leader>cx", function()
  require("treesitter-context").go_to_context(vim.v.count1)
end, { desc = "Go to context" })
