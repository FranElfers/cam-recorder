.PHONY: all build install clean uninstall

all: build

build:
	CGO_ENABLED=0 go build -o cam-recorder main.go

install: build
	install -d /usr/local/bin
	install -m 755 cam-recorder /usr/local/bin/
	install -d /etc/cam-recorder
	install -m 644 config.json /etc/cam-recorder/
	install -m 755 cam-recorder.initd /etc/init.d/cam-recorder
	rc-update add cam-recorder default

uninstall:
	rc-update del cam-recorder default
	rm -f /etc/init.d/cam-recorder
	rm -rf /etc/cam-recorder
	rm -f /usr/local/bin/cam-recorder

clean:
	rm -f cam-recorder
