# Prefix a red [sudo] when running elevated.
$isElevated = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($isElevated -and "${function:prompt}" -notmatch '__basePrompt') {
  $global:__basePrompt = ${function:prompt}
  function global:prompt {
    "`e[31m[sudo]`e[0m " + (& $global:__basePrompt)
  }
}

# Shows pending package changes in the prompt. Needs `make pkgsync`.
# The repo root is found from this file's real path (the profile is a symlink).
$profileFile = Get-Item $PSCommandPath
$env:PKGSYNC_ROOT ??= Split-Path (Split-Path (Split-Path ($profileFile.LinkTarget ?? $profileFile.FullName)))
$pkgsyncPs1 = Join-Path $env:PKGSYNC_ROOT 'tools\pkgsync\pkgsync.ps1'
if (Test-Path $pkgsyncPs1) { . $pkgsyncPs1 }

# Import-Module yarn-completion
# Import-Module posh-git

function reload {
  & $PROFILE.CurrentUserAllHosts
}

function config {
  code $PROFILE.CurrentUserAllHosts
}

function touch {
  param (
    [string]$path
  )
  New-Item -ItemType File -Path $path
}

function Set-Env {
  param (
    [Parameter(Mandatory)][string]$Name,
    [Parameter(Mandatory)][string]$Value
  )
  if (-not (Test-Path "Env:$Name")) {
    Set-Item -Path "Env:$Name" -Value $Value
  }
}

Set-Env CMAKE_GENERATOR "Ninja"
Set-Env BAT_PAGER "less -RF --mouse"

Remove-Item Function:\Set-Env

# Machine-local overrides.
$localProfile = Join-Path (Split-Path $PROFILE.CurrentUserAllHosts) 'profile.local.ps1'
if (Test-Path $localProfile) { . $localProfile }
