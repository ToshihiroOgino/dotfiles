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
cwd=$(jq -r '.cwd // empty' <<<"$input")
[ -n "$cwd" ] && cd "$cwd" 2>/dev/null

GCP_RE='gcloud|gsutil|(^|[^[:alnum:]_.-])bq([^[:alnum:]_-]|$)|google-cloud|cloudsdk|googleapis|google\.cloud|googleapiclient|firebase_admin|\.config/gcloud|application_default_credentials|GOOGLE_APPLICATION_CREDENTIALS|metadata\.google\.internal|169\.254\.169\.254|kubectl|\.kube/|KUBECONFIG|kubernetes'
OVERRIDE_RE='CLOUDSDK_|--account|--impersonate-service-account|--access-token-file|--configuration|--credential-file-override|--flags-file|KUBECONFIG|--kubeconfig|--as([^[:alnum:]-]|-group|-uid|$)|--token|--server|(^|[[:space:]])-s([[:space:]=]|$)|--username|--password|--client-key|--client-certificate|--insecure-skip-tls-verify'
TOOLS=' gcloud gcloud.py gsutil bq kubectl '
WRAPPERS=' env command exec nohup time sudo xargs nice timeout caffeinate '
RUNNERS=' bash sh zsh dash fish source . python python3 node deno bun ruby perl tsx ts-node npx uv uvx poetry pipenv go make just npm pnpm yarn task '
GCLOUD_BOOLS=' -q --quiet --log-http --no-user-output-enabled --user-output-enabled -h --help '
GSUTIL_BOOLS=' -m -q -D -DD -h '
BQ_BOOLS=' -q --quiet --headless --nouse_gce_service_account '
KUBECTL_BOOLS=' -h --help '

normalize() {
	tr -d "'\"\\\\" <<<"$1"
}

split_segments() {
	sed -E 's/(\$\(|<\(|>\(|`|&&|\|\||[;&|(){}])/\n/g' <<<"$1"
}

command_word() {
	local t
	for t in $1; do
		case "$t" in
		[A-Za-z_]*=*) continue ;;
		-*) continue ;;
		esac
		case "$WRAPPERS" in
		*" $t "*) continue ;;
		esac
		echo "$t"
		return
	done
}

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

check_tool_invocations() {
	local seg="$1" t
	set -- $seg
	while [ $# -gt 0 ]; do
		t="$1"
		shift
		case "${t##*/}" in
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
	done
}

scan_files() {
	local seg="$1" word="$2" t f
	local files=""
	case "$word" in
	make) files="Makefile makefile GNUmakefile" ;;
	just) files="justfile Justfile .justfile" ;;
	task) files="Taskfile.yml Taskfile.yaml" ;;
	npm | pnpm | yarn | bun) files="package.json" ;;
	esac
	for t in $seg; do
		files="$files $t"
	done
	for f in $files; do
		[ -f "$f" ] && [ -r "$f" ] || continue
		grep -qE "$GCP_RE" "$f" 2>/dev/null && deny "script references GCP tooling: $f"
	done
}

while IFS= read -r raw; do
	seg=$(normalize "$raw")
	word=$(command_word "$seg")
	case "$word" in
	\$* | eval) deny "dynamic command execution: $raw" ;;
	esac
done < <(sed -E 's/(&&|\|\||[;&|])/\n/g' <<<"$cmd")

grep -qE '\)[[:alnum:]_/]' <<<"$cmd" && grep -qE '\$\(|`' <<<"$cmd" &&
	deny "command substitution glued to a word: $cmd"
grep -qE '(base64|xxd|openssl)[^|]*\|[[:space:]]*(ba|z|da)?sh([^[:alnum:]]|$)' <<<"$cmd" &&
	deny "decoded content piped to shell: $cmd"

while IFS= read -r raw; do
	seg=$(normalize "$raw")
	[ -z "${seg// /}" ] && continue
	word=$(command_word "$seg")

	case "$RUNNERS" in
	*" ${word##*/} "*) scan_files "$seg" "${word##*/}" ;;
	esac
	case "$word" in
	*/*) scan_files "$word" "" ;;
	esac

	grep -qE "$GCP_RE" <<<"$seg" || continue
	grep -qE -- "$OVERRIDE_RE" <<<"$seg" && deny "credential/config override: $raw"
	case "$TOOLS" in
	*" ${word##*/} "*) check_tool_invocations "$seg" ;;
	*) deny "GCP/Kubernetes-related command outside gcloud/gsutil/bq/kubectl: $raw" ;;
	esac
done < <(split_segments "$cmd")

exit 0
