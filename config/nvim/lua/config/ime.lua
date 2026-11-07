-- Insert 以外では IME を無効にする (tools/ime)
local ime = vim.fn.expand("~/.local/bin/ime") .. (vim.fn.has("win32") == 1 and ".exe" or "")
if vim.fn.executable(ime) == 1 then
	local group = vim.api.nvim_create_augroup("ime_insert_only", { clear = true })
	local saved -- 最後に Insert を抜けたときの状態: "ja" / "en"
	local function run(...)
		vim.system({ ime, ... }, {}, function() end)
	end
	local function restore_enabled()
		if vim.b.ime_restore ~= nil then
			return vim.b.ime_restore
		end
		return vim.g.ime_restore == true
	end
	local function in_insert()
		return vim.fn.mode():match("^[iR]") ~= nil
	end

	vim.api.nvim_create_autocmd("InsertLeave", {
		group = group,
		callback = function()
			if not restore_enabled() then
				saved = nil
				run("set", "en")
				return
			end
			vim.system({ ime, "get" }, { text = true }, function(res)
				local state = vim.trim(res.stdout or "")
				if state == "ja" or state == "en" then
					saved = state
				end
				vim.schedule(function()
					if not in_insert() then
						run("set", "en")
					end
				end)
			end)
		end,
	})

	vim.api.nvim_create_autocmd("InsertEnter", {
		group = group,
		callback = function()
			if restore_enabled() and saved == "ja" then
				run("set", "ja")
			end
		end,
	})

	vim.api.nvim_create_autocmd({ "CmdlineEnter", "CmdlineLeave" }, {
		group = group,
		callback = function()
			if not in_insert() then
				run("set", "en")
			end
		end,
	})
end
