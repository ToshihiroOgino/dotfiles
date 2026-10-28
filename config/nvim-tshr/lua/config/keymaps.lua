vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Leave insert mode (fires InsertLeave)" })

vim.keymap.set("n", "<leader>ff", function() Snacks.picker.files() end, { desc = "Find files" })
vim.keymap.set("n", "<leader>/", function() Snacks.picker.grep() end, { desc = "Grep" })
vim.keymap.set("n", "<leader>fb", function() Snacks.picker.buffers() end, { desc = "Buffers" })
vim.keymap.set("n", "<leader>fr", function() Snacks.picker.recent() end, { desc = "Recent files" })
vim.keymap.set("n", "<leader>gg", function() Snacks.lazygit() end, { desc = "Lazygit" })
