$theme = "wopian"
oh-my-posh init pwsh --config $theme | Invoke-Expression

Import-Module yarn-completion
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

$env:CMAKE_GENERATOR = "Ninja"
