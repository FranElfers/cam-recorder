#!/bin/sh
# User-level installer. Only the optional OpenRC service asks for sudo.
set -e
BIN_DIR="$HOME/.local/bin"
DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/cam-recorder"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/cam-recorder"
STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/cam-recorder"
INITD=/etc/init.d/cam-recorder

if [ "$(id -u)" -eq 0 ]; then
	echo "Run this installer as your normal user, not as root." >&2
	exit 1
fi

PAYLOAD_LINE=$(awk '/^__PAYLOAD_BELOW__/ {print NR + 1; exit 0; }' "$0")
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
tail -n +"$PAYLOAD_LINE" "$0" | tar -xz -C "$TMP"

mkdir -p "$BIN_DIR" "$DATA_DIR/videos" "$DATA_DIR/hls" "$CONF_DIR" "$STATE_DIR"
install -m 755 "$TMP/cam-recorder" "$BIN_DIR/"
install -m 644 "$TMP/index.html" "$DATA_DIR/"
if [ ! -f "$CONF_DIR/config.json" ]; then
	sed "s|/var/lib/cam-recorder|$DATA_DIR|" "$TMP/config.json" > "$CONF_DIR/config.json"
	chmod 600 "$CONF_DIR/config.json"
fi

echo "Installation complete."
command -v ffmpeg >/dev/null || echo "WARNING: ffmpeg not found. Install it with: sudo apk add ffmpeg"
echo "Edit $CONF_DIR/config.json before starting."

printf "Enable the OpenRC service (start at boot; asks for your sudo password)? [y/N] "
read -r ANSWER || true
case "$ANSWER" in y|Y|yes|YES)
	# The service runs as this user and reads this user's configuration.
	cat > "$TMP/cam-recorder.initd" <<SERVICE
#!/sbin/openrc-run

name="cam-recorder"
description="Tapo C200 RTSP Recorder Service"
command="$BIN_DIR/cam-recorder"
command_user="$(id -u):$(id -g)"
command_background="yes"
pidfile="/run/\${RC_SVCNAME}.pid"
output_log="$STATE_DIR/cam-recorder.log"
error_log="$STATE_DIR/cam-recorder.err"
directory="$DATA_DIR"

export HOME="$HOME"

depend() {
    need net
    use logger
}
SERVICE
	sudo install -m 755 "$TMP/cam-recorder.initd" "$INITD"
	sudo rc-update add cam-recorder default
	echo "Service enabled. Manage it with:"
	echo "  sudo rc-service cam-recorder start|stop|restart|status"
	printf "Start the service now? Make sure config.json is edited first. [y/N] "
	read -r ANSWER || true
	case "$ANSWER" in y|Y|yes|YES) sudo rc-service cam-recorder restart ;; esac
	;;
*) echo "Service not enabled. Run it directly with: cd $DATA_DIR && $BIN_DIR/cam-recorder" ;;
esac
exit 0
__PAYLOAD_BELOW__
