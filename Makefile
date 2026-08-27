.PHONY: all build install clean uninstall

all: build

build:
	go build -o cam-recorder main.go

install: build
	cp cam-recorder /usr/local/bin/
	mkdir -p /etc/cam-recorder
	cp config.json /etc/cam-recorder/
	cp cam-recorder.initd /etc/init.d/cam-recorder
	chmod +x /etc/init.d/cam-recorder
	rc-update add cam-recorder default

uninstall:
	rc-update del cam-recorder default
	rm -f /etc/init.d/cam-recorder
	rm -rf /etc/cam-recorder
	rm -f /usr/local/bin/cam-recorder

clean:
	rm -f cam-recorder
