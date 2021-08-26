#!/bin/bash

# check to see if the script is being sourced or called
if [[ $0 != $BASH_SOURCE ]]; then
    SCRIPT_SOURCE=$BASH_SOURCE
else
    SCRIPT_SOURCE=$0
fi

THIS_DIR="$(dirname "$(realpath "$SCRIPT_SOURCE")")"
unset SCRIPT_SOURCE

{{ AppName }}_hist_start() { {{ AppExe }} history start "$1"; }
{{ AppName }}_hist_stop() { {{ AppExe }} history stop $?; }

source "${THIS_DIR}/bash-preexec.sh"
unset THIS_DIR

[[ " ${preexec_functions[*]} " =~ " {{ AppName }}_hist_start " ]] || \
    preexec_functions+=({{ AppName }}_hist_start)

[[ " ${precmd_functions[*]} " =~ " {{ AppName }}_hist_stop " ]] || \
    precmd_functions+=({{ AppName }}_hist_stop)

EXE_DIR="$(dirname "{{ AppExe }}")"
[[ ":${PATH}:" = *":${EXE_DIR}:"* ]] || PATH="${PATH}:${EXE_DIR}"
unset EXE_DIR

# wrapper to help with expansion of history lookups
aardy() {
    local args ch

    if [[ " $* " =~ " exec " ]]; then
        args=()
        for arg in "$@"; do
            ch=${arg:0:1}
            if [[ $ch = '^' ]]; then
                cmd=$({{ AppExe }} history get "$arg")
                args+=("$cmd")
            else
                args+=("$arg")
            fi
        done
    else
        args=("$@")
    fi

    {{ AppExe }} "${args[@]}"
}

export AARDY_INTEGRATED=true
