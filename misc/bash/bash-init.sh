
# PRUNEY_DIR refers to the current directory where pruney is installed
# This should contain the following
# - bash-init.sh, this file
# - bash-preexec.sh, the bash script which adds the bash-prexec and bash-precmd
# - pruney, the pruney cli binary
PRUNEY_DIR="$(dirname "$(realpath "$0")")"
PRUNEY_BIN=${PRUNEY_DIR}/pruney

preexec_append_history() { echo $1 | ${PRUNEY_BIN} history append-in; }

source "$PRUNEY_DIR/bash-preexec.sh"
preexec_functions+=(preexec_append_history)

PATH=$PATH:${PRUNEY_DIR}
