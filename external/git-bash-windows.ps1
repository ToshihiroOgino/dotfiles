# Point CLAUDE_CODE_GIT_BASH_PATH at Git Bash's bash.exe. Idempotent.
#   powershell -NoProfile -File external/git-bash-windows.ps1
# Claude Code needs it when git is not on PATH (scoop shims, installer without
# PATH integration). scoop wins over the Git for Windows installer. The value is
# bash.exe, not git-bash.exe (that one is the terminal launcher). HKCU only. New
# shells pick the change up.
$ErrorActionPreference = 'Stop'

$scoopRoot = if ($env:SCOOP) { $env:SCOOP } else { Join-Path $env:USERPROFILE 'scoop' }
$candidates = @(
  (Join-Path $scoopRoot 'apps\git\current\bin\bash.exe'),
  (Join-Path $env:ProgramFiles 'Git\bin\bash.exe')
)

$bash = $candidates | Where-Object { Test-Path -LiteralPath $_ -PathType Leaf } | Select-Object -First 1
if (-not $bash) {
  Write-Host 'skip     CLAUDE_CODE_GIT_BASH_PATH (Git Bash not found)'
  exit 0
}

if ([Environment]::GetEnvironmentVariable('CLAUDE_CODE_GIT_BASH_PATH', 'User') -eq $bash) {
  Write-Host "skip     CLAUDE_CODE_GIT_BASH_PATH=$bash (already set)"
  exit 0
}

# Also broadcasts the environment change to running programs.
[Environment]::SetEnvironmentVariable('CLAUDE_CODE_GIT_BASH_PATH', $bash, 'User')
Write-Host "set      CLAUDE_CODE_GIT_BASH_PATH=$bash"
