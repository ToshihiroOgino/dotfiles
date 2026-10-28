# Dot-source from the pwsh profile, after the prompt is set up: runs
# `pkgsync diff --throttle` in the background and shows packages/pending/summary in
# the prompt. Does nothing until `make pkgsync`.

$root = if ($env:PKGSYNC_ROOT) { $env:PKGSYNC_ROOT } else { Join-Path $HOME 'dotfiles' }
$exe = Join-Path $root 'tools\pkgsync\pkgsync.exe'
if (-not (Test-Path -LiteralPath $exe)) { return }

$env:PKGSYNC_ROOT = $root
Set-Alias -Name pkgsync -Value $exe -Scope Global
Start-Process -WindowStyle Hidden -FilePath $exe -ArgumentList 'diff','--throttle'

# Wrap the current prompt once, so that reloading the profile cannot nest it.
if ("${function:prompt}" -notmatch '__pkgsyncPrompt') {
  $global:__pkgsyncPrompt = ${function:prompt}
  $global:__pkgsyncSummary = Join-Path $root 'packages\pending\summary'
  function global:prompt {
    $p = & $global:__pkgsyncPrompt
    try { $s = [IO.File]::ReadAllText($global:__pkgsyncSummary).Trim() } catch { $s = '' }
    if ($s) { $p = "`e[33m[pkg$s]`e[0m" + $p }
    $p
  }
}
