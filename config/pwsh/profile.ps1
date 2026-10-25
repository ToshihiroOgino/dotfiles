$theme = "wopian"
oh-my-posh init pwsh --config $theme | Invoke-Expression

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

Remove-Item Function:\Set-Env

# Machine-local overrides.
$localProfile = Join-Path (Split-Path $PROFILE.CurrentUserAllHosts) 'profile.local.ps1'
if (Test-Path $localProfile) { . $localProfile }
