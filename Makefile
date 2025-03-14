all: debug build

debug:
	go build -o simulator_autostart_debug.exe simulator_autostart.go

build:
	go build -ldflags -H=windowsgui simulator_autostart.go

test:
	go test -v ./...