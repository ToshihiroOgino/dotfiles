-- Open images / PDFs in the OS default app instead of showing binary in a buffer.
local pattern = { "*.png", "*.jpg", "*.jpeg", "*.gif", "*.webp", "*.bmp", "*.ico", "*.pdf" }

return {
  {
    "folke/snacks.nvim",
    init = function()
      vim.api.nvim_create_autocmd("BufReadCmd", {
        group = vim.api.nvim_create_augroup("open_external", { clear = true }),
        pattern = pattern,
        callback = function(ev)
          vim.ui.open(vim.fn.fnamemodify(ev.file, ":p"))
          vim.schedule(function()
            if vim.api.nvim_buf_is_valid(ev.buf) then
              pcall(vim.api.nvim_buf_delete, ev.buf, { force = true })
            end
          end)
        end,
      })
    end,
    keys = {
      {
        "<leader>fo",
        function()
          vim.ui.open(vim.fn.expand("%:p"))
        end,
        desc = "Open with default app",
      },
    },
  },
}
