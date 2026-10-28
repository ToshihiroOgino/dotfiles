# Put a directory on the user PATH through DOTFILES_PATH. Idempotent.
#   powershell -NoProfile -File external/path-windows.ps1 -Dir <dir> -Root <parent>
# The user Path only ever gains one entry, %DOTFILES_PATH%; the directories live
# in the DOTFILES_PATH user variable. Entries under -Root (older versions of the
# same tool) are replaced by -Dir. HKCU only. New shells pick the change up.
param(
  [Parameter(Mandatory)][string]$Dir,
  [Parameter(Mandatory)][string]$Root
)
$ErrorActionPreference = 'Stop'

$entries = @([Environment]::GetEnvironmentVariable('DOTFILES_PATH', 'User') -split ';' | Where-Object { $_ })
$prefix = $Root.TrimEnd('\') + '\'
$entries = @($entries | Where-Object { -not $_.StartsWith($prefix, [StringComparison]::OrdinalIgnoreCase) })
$entries = @($Dir) + $entries

# Read the raw value: expanding it would bake %VAR% references into the Path.
$key = Get-Item 'HKCU:\Environment'
$path = $key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
if (($path -split ';') -notcontains '%DOTFILES_PATH%') {
  $path = (($path -split ';' | Where-Object { $_ }) + '%DOTFILES_PATH%') -join ';'
  # REG_EXPAND_SZ, or %VAR% in Path is no longer expanded.
  Set-ItemProperty -Path 'HKCU:\Environment' -Name Path -Value $path -Type ExpandString
  Write-Host 'update   user Path (+%DOTFILES_PATH%)'
}

# Also broadcasts the environment change to running programs.
[Environment]::SetEnvironmentVariable('DOTFILES_PATH', ($entries -join ';'), 'User')
Write-Host "set      DOTFILES_PATH=$($entries -join ';')"
