# Source from zshrc: runs `pkgsync diff --throttle` in the background and shows
# packages/pending/summary in the prompt. Does nothing until `make pkgsync`.

_pkgsync_root=${PKGSYNC_ROOT:-$HOME/dotfiles}
export PKGSYNC_ROOT=$_pkgsync_root
_pkgsync_bin=$_pkgsync_root/tools/pkgsync/pkgsync
[[ -x $_pkgsync_bin ]] || return 0

# The subshell keeps zsh from printing a job number.
( "$_pkgsync_bin" diff --throttle >/dev/null 2>&1 & )

# PROMPT is edited here so that a theme loaded later cannot overwrite it.
_pkgsync_precmd() {
  [[ $PROMPT == *_pkgsync_seg* ]] || PROMPT='${_pkgsync_seg}'$PROMPT
  local f=$_pkgsync_root/packages/pending/summary
  if [[ -r $f ]]; then
    _pkgsync_seg="%F{yellow}[pkg$(<$f)]%f"
  else
    _pkgsync_seg=
  fi
}
autoload -Uz add-zsh-hook
add-zsh-hook precmd _pkgsync_precmd
setopt prompt_subst
