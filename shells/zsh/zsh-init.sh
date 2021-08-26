#!/bin/zsh

function {{ AppName }}_hist_start() {
    local cmd=${1%% *}
    # attempt to "expand" a command if it's an alias...0
    cmd=$(type $cmd | sed -n 's/.* is an alias for \(.*\)/\1/p')
    # so we can ignore anything aliased as an aardy history command...
    [[ "$cmd" = 'aardy h'* ]] || {{ AppExe }} history start "$1"
}

function {{ AppName }}_hist_stop() { {{ AppExe }} history stop $?; }

autoload -Uz add-zsh-hook
add-zsh-hook preexec {{ AppName }}_hist_start
add-zsh-hook precmd {{ AppName }}_hist_stop

EXE_DIR="$(dirname "{{ AppExe }}")"
[[ ":$PATH:" = *":${EXE_DIR}:"* ]] || PATH="${PATH}:${EXE_DIR}"
unset EXE_DIR

export AARDY_INTEGRATED=true
