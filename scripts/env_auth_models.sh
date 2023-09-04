#!/usr/bin/env bash

set -euo pipefail

: "${ENV_FILE:=$(realpath "$(dirname "$0" )/.." )/.env}"

function _hack {
	git submodule update --init || return 1
	local -r HACK_DIR="$(dirname "$ENV_FILE" )"
	( cat "$HACK_DIR/README.md" - | tail -n+2 1>&2 ) << EOF

--- N.B. includes whisper.cpp in: $HACK_DIR # see READMEs and CMakeLists.txt

Georgi Gerganov is the author of https://github.com/ggerganov/whisper.cpp
New macOS requires cmake: brew install --formula cmake # and then:
DIR=$HACK_DIR/whisper.cpp/build; mkdir -p \$DIR && cd \$DIR

# edit ../CMakeLists.txt top line: cmake_minimum_required(VERSION 3.5)
cmake .. && make && make install # to install whisper.cpp libraries

EOF
}

function _main {
	echo 1>&2 "--- $(date): $0 $*"
	_hack || return 1

	if [[ -r "$ENV_FILE" ]]
	then echo 1>&2 "--- $ENV_FILE # contains secrets"
		cut -d '=' -f1 < "$ENV_FILE" | grep _SECRET
		grep -v SECRET < "$ENV_FILE" || true
		echo 1>&2 "---"
	else echo 1>&2 "--- missing .env file: $ENV_FILE"
	fi
}

_main "$@"
