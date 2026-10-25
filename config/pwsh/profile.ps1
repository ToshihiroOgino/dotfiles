$theme = "wopian"
oh-my-posh init pwsh --config $theme | Invoke-Expression

# Prefix the prompt with a red [sudo] when running elevated.
$isElevated = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($isElevated) {
  $global:__ompPrompt = ${function:prompt}
  function global:prompt {
    "`e[31m[sudo]`e[0m " + (& $global:__ompPrompt)
  }
}

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
