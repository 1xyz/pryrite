#!/bin/bash

# check to see if the script is being sourced or called
if [[ $0 != $BASH_SOURCE ]]; then
    SCRIPT_SOURCE=$BASH_SOURCE
else
    SCRIPT_SOURCE=$0
fi

THIS_DIR="$(dirname "$(realpath "$SCRIPT_SOURCE")")"
unset SCRIPT_SOURCE

{{ AppName }}_hist_start() {
    local cmd=${1%% *}
    # attempt to "expand" a command if it's an alias...0
    cmd=$(type $cmd 2>/dev/null | sed -n 's/.*`\([^'\'']*\).*/\1/p')
    # so we can ignore anything aliased as an aardy history command...
    [[ "$cmd" = 'aardy h'* ]] || {{ AppExe }} history start "$1"
}

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

export AARDY_INTEGRATED=true
