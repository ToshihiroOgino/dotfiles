-- ここを書き換えるとテーマが切り替わる
--
-- 寒色系
--   "tokyonight"            青〜紫 (style: night / storm / moon / day)
--   "nord"                  くすんだ北欧ブルー
--   "kanagawa"              藍色ベースの和風 (kanagawa-wave / kanagawa-dragon)
--   "catppuccin-frappe"     落ち着いた青紫 (catppuccin-macchiato / catppuccin-mocha もあり)
--   "nightfox"              青系 (nordfox / duskfox / terafox / carbonfox もあり)
-- 暖色系
--   "gruvbox"               ブラウン〜オレンジ
local theme = "tokyonight"

return {
  {
    "folke/tokyonight.nvim",
    lazy = true,
    opts = {
      style = "night",
      transparent = false,
    },
  },
  {
    "gbprod/nord.nvim",
    lazy = true,
  },
  {
    "rebelot/kanagawa.nvim",
    lazy = true,
  },
  {
    "catppuccin/nvim",
    name = "catppuccin",
    lazy = true,
  },
  {
    "EdenEast/nightfox.nvim",
    lazy = true,
  },
  {
    "ellisonleao/gruvbox.nvim",
    lazy = true,
    opts = {
      contrast = "hard",
      transparent_mode = false,
    },
  },
  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = theme,
    },
  },
}
