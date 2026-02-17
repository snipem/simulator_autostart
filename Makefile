all: debug build

debug:
	go build -o simulator_autostart_debug.exe ./cmd/console

build:
	go build -ldflags "-s" -o simulator_autostart.exe ./cmd/console

build-win:
	go build -ldflags "-s -H=windowsgui" -o simulator_autostart_win.exe ./cmd/win

test:
	go test -v ./...
