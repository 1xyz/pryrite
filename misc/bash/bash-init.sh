# check to see if the script is being sourced or called
if [[ $0 != $BASH_SOURCE ]]
then
    SCRIPT_SOURCE=$BASH_SOURCE
else
    SCRIPT_SOURCE=$0
fi

# AARD_DIR refers to the current directory where the CLI is installed
# This should contain the following
# - bash-init.sh, this file
# - bash-preexec.sh, the bash script which adds the bash-prexec and bash-precmd
# - aard, the  cli binary
AARD_DIR="$(dirname "$(realpath "$SCRIPT_SOURCE")")"
AARD_BIN=${AARD_DIR}/aard

preexec_append_history() { echo $1 | ${AARD_BIN} history append --stdin; }

source "$AARD_DIR/bash-preexec.sh"
preexec_functions+=(preexec_append_history)

# add $AARD_DIR to PATH
[[ ":$PATH:" != *":${AARD_DIR}:"* ]] && PATH="${PATH}:${AARD_DIR}"
