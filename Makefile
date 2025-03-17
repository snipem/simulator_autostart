all: debug build

debug:
	go build -o simulator_autostart_debug.exe simulator_autostart.go

build:
	go build -ldflags "-s -H=windowsgui" simulator_autostart.go

test:
	go test -v ./...