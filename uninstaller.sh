#!/bin/sh
# User-level uninstaller. Only the OpenRC service removal asks for sudo.
set -e
BIN_DIR="$HOME/.local/bin"
DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/cam-recorder"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/cam-recorder"
INITD=/etc/init.d/cam-recorder

if [ -f "$INITD" ]; then
	printf "Stop and disable the OpenRC service (asks for your sudo password)? [Y/n] "
	read -r ANSWER
	case "$ANSWER" in
		n|N|no|NO) echo "Service kept. Its binary is removed below, so it will fail to start." ;;
		*)
			sudo rc-service cam-recorder stop || true
			sudo rc-update del cam-recorder default || true
			sudo rm -f "$INITD" ;;
	esac
fi

rm -f "$BIN_DIR/cam-recorder" "$DATA_DIR/index.html"
echo "Removed. Configuration ($CONF_DIR) and recordings ($DATA_DIR) were kept."
