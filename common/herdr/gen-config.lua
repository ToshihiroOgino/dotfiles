#!/usr/bin/env lua
-- config.base.toml と OS 別オーバーレイをマージして generated/ に書き出す。
-- herdr が設定ファイルの include に対応していないための代替手段。
--
-- Lua 5.1 互換なので lua / lua5.4 / `nvim -l` のどれでも動く。外部ライブラリ不要。
--
--   lua gen-config.lua                 全OS分を生成
--   lua gen-config.lua windows linux   指定したOSだけ生成
--   lua gen-config.lua --clean         生成物を削除
--
-- マージ規則:
--   * 同じ [table] 内の同名キーはオーバーレイ側が勝つ (出現位置は base のまま)
--   * オーバーレイにしかないキー / テーブルは末尾に追加される
--   * [[array.of.tables]] はオーバーレイが定義したらその path の base 側を全置換
--   * キーの直前のコメント行と空行はそのキーに追随する
--   * ファイル先頭のコメント塊はそのファイル自身の説明とみなして生成物には出さない

local OSES = { "windows", "linux", "macos" }

local BASE = "config.base.toml"
local OUT = "generated/config.%s.toml"

local WINDOWS = package.config:sub(1, 1) == "\\"

local function die(fmt, ...)
  io.stderr:write("gen-config: " .. string.format(fmt, ...) .. "\n")
  os.exit(1)
end

local function read_file(path)
  local f = io.open(path, "rb")
  if not f then
    return nil
  end
  local text = f:read("*a")
  f:close()
  return text
end

local function mkdir(dir)
  if WINDOWS then
    os.execute('mkdir "' .. dir:gsub("/", "\\") .. '"')
  else
    os.execute('mkdir -p "' .. dir .. '"')
  end
end

local function write_file(path, text)
  -- symlink の実体を保つため、消して作り直さず truncate して上書きする。
  local f = io.open(path, "wb")
  if not f then
    local dir = path:match("^(.*)/[^/]*$")
    if dir then
      mkdir(dir)
      f = io.open(path, "wb")
    end
  end
  if not f then
    die("cannot write %s", path)
  end
  f:write(text)
  f:close()
end

-- 文字列リテラルと行コメントを無視して [ { の深さを更新する。
-- 複数行にまたがる値 (配列やインラインテーブル) の終端判定に使う。
local function track_depth(line, depth)
  local quote = nil
  local i = 1
  while i <= #line do
    local c = line:sub(i, i)
    if quote then
      if c == "\\" and quote == '"' then
        i = i + 1
      elseif c == quote then
        quote = nil
      end
    elseif c == '"' or c == "'" then
      quote = c
    elseif c == "#" then
      break
    elseif c == "[" or c == "{" then
      depth = depth + 1
    elseif c == "]" or c == "}" then
      depth = depth - 1
    end
    i = i + 1
  end
  return depth
end

local function split_lines(text)
  local lines = {}
  for line in (text .. "\n"):gmatch("([^\n]*)\n") do
    lines[#lines + 1] = (line:gsub("\r$", ""))
  end
  -- gmatch で末尾に空行が 1 本余るので落とす
  if lines[#lines] == "" then
    lines[#lines] = nil
  end
  return lines
end

local function new_holder()
  return { keys = {}, entries = {} }
end

-- doc = { order = {path...}, sections = { [path] = section } }
-- section = { path, kind = "table"|"array", header, lead, keys, entries, items }
local function parse(text, where)
  local lines = split_lines(text)
  local doc = { order = {}, sections = {} }

  local function section(path, kind, header, lead)
    local s = doc.sections[path]
    if not s then
      s = new_holder()
      s.path, s.kind, s.header, s.lead, s.items = path, kind, header, lead or {}, {}
      doc.sections[path] = s
      doc.order[#doc.order + 1] = path
    elseif s.kind ~= kind then
      die("%s: [%s] is declared both as a table and an array of tables", where, path)
    end
    return s
  end

  local root = section("", "table", nil, {})
  local current = root -- キーの行き先 (テーブル or 配列テーブルの要素)
  local lead = {}
  local i = 1

  -- 先頭のコメント塊 + 続く空行はファイル自身の説明なので読み捨てる。
  while lines[i] and lines[i]:match("^%s*#") do
    i = i + 1
  end
  if i > 1 then
    while lines[i] and lines[i]:match("^%s*$") do
      i = i + 1
    end
  end

  while i <= #lines do
    local line = lines[i]
    local trimmed = line:match("^%s*(.-)%s*$")

    if trimmed == "" or trimmed:sub(1, 1) == "#" then
      lead[#lead + 1] = line
      i = i + 1
    else
      local array_path = trimmed:match("^%[%[%s*(.-)%s*%]%]$")
      local table_path = (not array_path) and trimmed:match("^%[%s*(.-)%s*%]$") or nil

      if array_path then
        local s = section(array_path, "array", "[[" .. array_path .. "]]", lead)
        local item = new_holder()
        item.lead = lead
        s.items[#s.items + 1] = item
        current = item
        lead = {}
        i = i + 1
      elseif table_path then
        if doc.sections[table_path] and doc.sections[table_path].seen then
          die("%s: [%s] appears twice", where, table_path)
        end
        local s = section(table_path, "table", "[" .. table_path .. "]", lead)
        s.seen = true
        current = s
        lead = {}
        i = i + 1
      else
        local key = trimmed:match("^(.-)%s*=")
        if not key or key == "" then
          die("%s: cannot parse line %d: %s", where, i, line)
        end
        local body = { line }
        local depth = track_depth(line, 0)
        while depth > 0 and i < #lines do
          i = i + 1
          body[#body + 1] = lines[i]
          depth = track_depth(lines[i], depth)
        end
        if depth > 0 then
          die("%s: unterminated value for '%s'", where, key)
        end
        i = i + 1
        if current.entries[key] then
          die("%s: duplicate key '%s'", where, key)
        end
        current.keys[#current.keys + 1] = key
        current.entries[key] = { lines = body, lead = lead }
        lead = {}
      end
    end
  end

  return doc
end

local function merge_holder(base, overlay)
  for _, key in ipairs(overlay.keys) do
    if not base.entries[key] then
      base.keys[#base.keys + 1] = key
    end
    base.entries[key] = overlay.entries[key]
  end
end

local function merge(base, overlay, where)
  for _, path in ipairs(overlay.order) do
    local over = overlay.sections[path]
    local into = base.sections[path]
    if not into then
      base.sections[path] = over
      base.order[#base.order + 1] = path
    elseif into.kind ~= over.kind then
      die("%s: [%s] kind differs from the base file", where, path)
    elseif over.kind == "array" then
      into.items = over.items
    else
      merge_holder(into, over)
    end
  end
  return base
end

local function emit_holder(out, holder)
  for _, key in ipairs(holder.keys) do
    local entry = holder.entries[key]
    for _, line in ipairs(entry.lead) do
      out[#out + 1] = line
    end
    for _, line in ipairs(entry.lines) do
      out[#out + 1] = line
    end
  end
end

local function render(doc, header)
  local out = {}
  for _, line in ipairs(header) do
    out[#out + 1] = line
  end

  -- セクション見出しの手前は必ず 1 行空ける。
  -- lead 末尾のコメントは見出しに付いたままにする。
  local function emit_header(lead, text)
    if #out > 0 and not out[#out]:match("^%s*$") then
      local first = lead[1]
      if first == nil or not first:match("^%s*$") then
        out[#out + 1] = ""
      end
    end
    for _, line in ipairs(lead) do
      out[#out + 1] = line
    end
    out[#out + 1] = text
  end

  for _, path in ipairs(doc.order) do
    local s = doc.sections[path]
    if s.kind == "array" then
      for _, item in ipairs(s.items) do
        emit_header(item.lead, s.header)
        emit_holder(out, item)
      end
    else
      if path ~= "" then
        emit_header(s.lead, s.header)
      end
      emit_holder(out, s)
    end
  end

  while #out > 0 and out[#out]:match("^%s*$") do
    out[#out] = nil
  end
  return table.concat(out, "\n") .. "\n"
end

local function generate(dir, os_name)
  local overlay_name = os_name .. ".toml"
  local out_rel = string.format(OUT, os_name)

  local base_text = read_file(dir .. "/" .. BASE)
  if not base_text then
    die("missing base file %s", BASE)
  end

  local sources = { BASE }
  local doc = parse(base_text, BASE)

  local overlay_text = read_file(dir .. "/" .. overlay_name)
  if overlay_text then
    doc = merge(doc, parse(overlay_text, overlay_name), overlay_name)
    sources[#sources + 1] = overlay_name
  end

  local header = {
    "# GENERATED FILE - DO NOT EDIT",
    "# source: " .. table.concat(sources, " + "),
    "# regen:  make " .. os_name,
    "",
  }

  write_file(dir .. "/" .. out_rel, render(doc, header))
  print("generated " .. out_rel)
  if not overlay_text then
    print(string.format("  note: no %s overlay, wrote base only", overlay_name))
  end
end

local function clean(dir, os_name)
  local out_rel = string.format(OUT, os_name)
  if os.remove(dir .. "/" .. out_rel) then
    print("removed " .. out_rel)
  end
end

local function main()
  local script = (arg and arg[0] or "gen-config.lua"):gsub("\\", "/")
  local dir = script:match("^(.*)/[^/]+$") or "."

  local wanted, action = {}, generate
  for _, a in ipairs(arg or {}) do
    if a == "--clean" then
      action = clean
    elseif a:sub(1, 2) == "--" then
      die("unknown option %s", a)
    else
      local known = false
      for _, name in ipairs(OSES) do
        known = known or name == a
      end
      if not known then
        die("unknown os '%s' (expected: %s)", a, table.concat(OSES, ", "))
      end
      wanted[#wanted + 1] = a
    end
  end
  if #wanted == 0 then
    wanted = OSES
  end

  for _, os_name in ipairs(wanted) do
    action(dir, os_name)
  end
end

main()
