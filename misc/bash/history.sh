#!/bin/bash

ht() { HISTTIMEFORMAT='%F_%T  ' history; } # history with time stamps
h() { history; } # history with no timestamps

HISTFILE=$HOME/.bash_history

if [[ -e $HISTFILE ]]; then
	echo "History file is $HISTFILE"
	set -o history
	h
else
	echo "History is disabled"
fi
