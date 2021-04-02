# check to see if the script is being sourced or called
if [[ $0 != $BASH_SOURCE ]]
then
    SCRIPT_SOURCE=$BASH_SOURCE
else
    SCRIPT_SOURCE=$0
fi

# PRUNEY_DIR refers to the current directory where pruney is installed
# This should contain the following
# - bash-init.sh, this file
# - bash-preexec.sh, the bash script which adds the bash-prexec and bash-precmd
# - pruney, the pruney cli binary
PRUNEY_DIR="$(dirname "$(realpath "$SCRIPT_SOURCE")")"
PRUNEY_BIN=${PRUNEY_DIR}/pruney

preexec_append_history() { echo $1 | ${PRUNEY_BIN} history append --stdin; }

source "$PRUNEY_DIR/bash-preexec.sh"
preexec_functions+=(preexec_append_history)

# add $PRUNEY_DIR to PATH
[[ ":$PATH:" != *":${PRUNEY_DIR}:"* ]] && PATH="${PATH}:${PRUNEY_DIR}"
