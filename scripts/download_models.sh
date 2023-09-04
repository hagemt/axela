#!/usr/bin/env bash
# USAGE: git submodule update --init && ./scripts/download_models.sh

set -euo pipefail

: "${PULL_DIR:=$HOME/.whisper/models}"

: "${TOOL_DIR:=$(realpath "$(dirname "$0" )/../whisper.cpp" )}"
: "${MAIN_DIR:=$TOOL_DIR/bindings/go/examples/go-model-download}"

function _pull {
	mkdir -p "$PULL_DIR" || return 1
	[[ -d "$MAIN_DIR" ]] || return 2
	if pushd >/dev/null "$MAIN_DIR"
	then trap 'popd >/dev/null' EXIT
		exec "${GO:-$(command -v go)}" run . -out="$PULL_DIR" "$@"
	fi
}

_pull "$@"
