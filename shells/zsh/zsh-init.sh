#!/bin/zsh

function {{ AppName }}_hist_start() { {{ AppExe }} history start "$1"; }
function {{ AppName }}_hist_stop() { {{ AppExe }} history stop $?; }

autoload -Uz add-zsh-hook
add-zsh-hook preexec {{ AppName }}_hist_start
add-zsh-hook precmd {{ AppName }}_hist_stop

EXE_DIR="$(dirname "{{ AppExe }}")"
[[ ":$PATH:" = *":${EXE_DIR}:"* ]] || PATH="${PATH}:${EXE_DIR}"
unset EXE_DIR

# wrapper to help with expansion of history lookups
# FIXME: share the wrapper between zsh and bash!!
function aardy() {
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

AARDY_INTEGRATED=true
