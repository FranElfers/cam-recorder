.PHONY: all build clean installer

all: build

build:
	CGO_ENABLED=0 go build -o cam-recorder main.go

clean:
	rm -f cam-recorder cam-recorder-install.sh cam-recorder-uninstall.sh payload.tar.gz

installer: build
	tar -czf payload.tar.gz cam-recorder config.json index.html
	cat installer-head.sh payload.tar.gz > cam-recorder-install.sh
	chmod +x cam-recorder-install.sh
	rm payload.tar.gz
	cp uninstaller.sh cam-recorder-uninstall.sh
	chmod +x cam-recorder-uninstall.sh
