.PHONY: all build install clean uninstall installer

all: build

build:
	CGO_ENABLED=0 go build -o cam-recorder main.go

install: build
	install -d /usr/local/bin
	install -m 755 cam-recorder /usr/local/bin/
	install -d /etc/cam-recorder
	if [ ! -f /etc/cam-recorder/config.json ]; then install -m 644 config.json /etc/cam-recorder/; fi
	install -m 644 index.html /etc/cam-recorder/
	install -m 755 cam-recorder.initd /etc/init.d/cam-recorder
	rc-update add cam-recorder default
	rc-service cam-recorder restart || true

uninstall:
	rc-update del cam-recorder default
	rm -f /etc/init.d/cam-recorder
	rm -rf /etc/cam-recorder
	rm -f /usr/local/bin/cam-recorder

clean:
	rm -f cam-recorder cam-recorder-install.sh payload.tar.gz

installer: build
	tar -czf payload.tar.gz cam-recorder config.json index.html cam-recorder.initd
	echo '#!/bin/sh' > cam-recorder-install.sh
	echo 'PAYLOAD_LINE=$$(awk "/^__PAYLOAD_BELOW__/ {print NR + 1; exit 0; }" $$0)' >> cam-recorder-install.sh
	echo 'mkdir -p /tmp/cam-recorder-installer' >> cam-recorder-install.sh
	echo 'tail -n +$$PAYLOAD_LINE $$0 | tar -xz -C /tmp/cam-recorder-installer' >> cam-recorder-install.sh
	echo 'cd /tmp/cam-recorder-installer' >> cam-recorder-install.sh
	echo 'install -d /usr/local/bin' >> cam-recorder-install.sh
	echo 'install -m 755 cam-recorder /usr/local/bin/' >> cam-recorder-install.sh
	echo 'install -d /etc/cam-recorder' >> cam-recorder-install.sh
	echo 'if [ ! -f /etc/cam-recorder/config.json ]; then install -m 644 config.json /etc/cam-recorder/; fi' >> cam-recorder-install.sh
	echo 'install -m 644 index.html /etc/cam-recorder/' >> cam-recorder-install.sh
	echo 'install -m 755 cam-recorder.initd /etc/init.d/cam-recorder' >> cam-recorder-install.sh
	echo 'rc-update add cam-recorder default' >> cam-recorder-install.sh
	echo 'rc-service cam-recorder restart || true' >> cam-recorder-install.sh
	echo 'cd /' >> cam-recorder-install.sh
	echo 'rm -rf /tmp/cam-recorder-installer' >> cam-recorder-install.sh
	echo 'echo "Installation complete."' >> cam-recorder-install.sh
	echo 'exit 0' >> cam-recorder-install.sh
	echo '__PAYLOAD_BELOW__' >> cam-recorder-install.sh
	cat payload.tar.gz >> cam-recorder-install.sh
	chmod +x cam-recorder-install.sh
	rm payload.tar.gz
