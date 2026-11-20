#!/usr/bin/env bash
set -uf

CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/claude-cloud-guard/readonly-commands.txt"

if [ "${1:-}" = "--build" ]; then
	set -o pipefail
	mkdir -p "$(dirname "$CACHE")"
	if gcloud meta list-commands |
		grep -E ' (list|list-[a-z-]+|describe|get-iam-policy|read|tail|ls|cat|du|search-all-[a-z-]+)$|^gcloud (version|info|config get|config get-value)$' |
		sed 's/^gcloud //' >"$CACHE.tmp" && [ -s "$CACHE.tmp" ]; then
		mv "$CACHE.tmp" "$CACHE"
		wc -l "$CACHE"
		exit 0
	fi
	rm -f "$CACHE.tmp"
	echo "claude-cloud-guard: failed to build $CACHE" >&2
	exit 1
fi

deny() {
	echo "claude-cloud-guard: $1" >&2
	exit 2
}

command -v jq >/dev/null 2>&1 || deny "jq not found"
input=$(cat)
cmd=$(jq -er '.tool_input.command | strings' <<<"$input") || deny "failed to parse hook input"
dir=$(jq -r '.cwd // empty' <<<"$input")
dir="${dir:-$PWD}"

CONFIG_RE='\.config/gcloud|\.kube([/[:space:]]|$)|application_default_credentials'
READERS=' cat ls head tail less grep rg jq stat wc diff '
WRAPPERS=' env command exec nohup time timeout nice xargs sudo caffeinate bash sh zsh dash '
GCLOUD_BOOLS=' -q --quiet --log-http --no-user-output-enabled --user-output-enabled -h --help '
GSUTIL_BOOLS=' -m -q -D -DD -h '
BQ_BOOLS=' -q --quiet --headless --nouse_gce_service_account '
KUBECTL_BOOLS=' -h --help '

positionals() {
	local bools="$1" t
	shift
	while [ $# -gt 0 ]; do
		t="$1"
		shift
		case "$t" in
		-*=*) ;;
		-*)
			case "$bools" in
			*" $t "*) ;;
			*) [ $# -gt 0 ] && shift ;;
			esac
			;;
		*) echo "$t" ;;
		esac
	done
}

nth() {
	local n="$1"
	shift
	positionals "$@" | sed -n "${n}p"
}

is_readonly_gcloud() {
	local path="" t
	for t in $(positionals "$GCLOUD_BOOLS" "$@"); do
		path="${path:+$path }$t"
		grep -qxF "$path" "$CACHE" && return 0
	done
	return 1
}

is_readonly_kubectl() {
	case "$(nth 1 "$KUBECTL_BOOLS" "$@")" in
	get | describe)
		grep -qiE '(^|[[:space:],/])secrets?([[:space:],/.]|$)' <<<"$*" && return 1
		return 0
		;;
	logs | top | explain | events | api-resources | api-versions | version | cluster-info) return 0 ;;
	config)
		case "$(nth 2 "$KUBECTL_BOOLS" "$@")" in
		view) grep -qE -- '--raw' <<<"$*" && return 1 || return 0 ;;
		get-contexts | current-context | get-clusters) return 0 ;;
		esac
		;;
	auth)
		case "$(nth 2 "$KUBECTL_BOOLS" "$@")" in
		can-i | whoami) return 0 ;;
		esac
		;;
	esac
	return 1
}

check_segment() {
	local seg="$1" word=""
	set -- $seg
	while [ $# -gt 0 ]; do
		case "$1" in
		[A-Za-z_]*=* | -* | [0-9]*) ;;
		*)
			case "$WRAPPERS" in
			*" ${1##*/} "*) ;;
			*)
				word="${1##*/}"
				shift
				break
				;;
			esac
			;;
		esac
		shift
	done
	case "$word" in
	gcloud | gcloud.py)
		[ -s "$CACHE" ] || deny "cache missing; run: $0 --build"
		is_readonly_gcloud "$@" || deny "not a read-only gcloud command: $seg"
		;;
	gsutil)
		case "$(nth 1 "$GSUTIL_BOOLS" "$@")" in
		ls | cat | du | stat | hash | version) ;;
		*) deny "not a read-only gsutil command: $seg" ;;
		esac
		;;
	bq)
		case "$(nth 1 "$BQ_BOOLS" "$@")" in
		ls | show | head | version) ;;
		*) deny "not a read-only bq command: $seg" ;;
		esac
		;;
	kubectl)
		is_readonly_kubectl "$@" || deny "not a read-only kubectl command: $seg"
		;;
	esac
	case "$word" in
	cd | pushd)
		case "${1:-}" in
		"") dir="$HOME" ;;
		/* | \~* | \$*) dir="$1" ;;
		*) dir="$dir/$1" ;;
		esac
		;;
	esac
	local text="$seg $dir" t p w=0
	for t in "$@"; do
		case "$t" in
		*'>')
			w=1
			continue
			;;
		*'>'*)
			t="${t##*>}"
			w=1
			;;
		esac
		case "$t" in
		-*) p="" ;;
		/* | \~* | \$*) p="$t" ;;
		*) p="$dir/$t" ;;
		esac
		text="$text $p"
		[ "$w" = 1 ] && grep -qE "$CONFIG_RE" <<<"$p" && deny "write to cloud config: $seg"
		w=0
	done
	grep -qE "$CONFIG_RE" <<<"$text" || return 0
	grep -qE ">[[:space:]]*[^[:space:]]*($CONFIG_RE)" <<<"$seg" && deny "write to cloud config: $seg"
	case "$READERS gcloud gcloud.py gsutil bq kubectl " in
	*" $word "*) ;;
	*) deny "cloud config path used by non-reader: $seg" ;;
	esac
}

while IFS= read -r seg; do
	[ -z "${seg// /}" ] && continue
	check_segment "$seg"
done < <(tr -d "'\"\\\\" <<<"$cmd" | sed -E 's/(\$\(|<\(|>\(|`|&&|\|\||[;&|(){}])/\n/g')

exit 0
