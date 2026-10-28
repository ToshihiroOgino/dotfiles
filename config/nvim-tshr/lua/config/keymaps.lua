-- Ctrl-C は InsertLeave を発火しないので、Esc として扱う (ime.lua の退避もこれに乗る)
vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Leave insert mode (fires InsertLeave)" })
