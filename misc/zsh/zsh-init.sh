#!/bin/zsh

SCRIPT_SOURCE=$0
AARD_DIR="$(dirname "$(realpath "$SCRIPT_SOURCE")")"
AARD_BIN=${AARD_DIR}/aard

function aard_append_history() { echo $1 | ${AARD_BIN} history append --stdin; }

autoload -Uz add-zsh-hook
add-zsh-hook zshaddhistory aard_append_history

[[ ":$PATH:" != *":${AARD_DIR}:"* ]] && PATH="${PATH}:${AARD_DIR}"