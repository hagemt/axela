#!/usr/bin/env bash

set -euo pipefail

: "${ENV_FILE:=$(dirname "$0" )/../dev.env}"

function _main {
	echo 1>&2 "--- $(date): $0 $* # .env from: $ENV_FILE"

	# shellcheck source=../dev.env
	source "$ENV_FILE" || return 1

	local -r RUN="${1:-}"
	shift 2>/dev/null || true
	[[ -x "$RUN" ]] && exec "$RUN" "$@"
}

_main "$@"
