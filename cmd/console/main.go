package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"simulator_autostart/lib/autostart"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

var tagColors = map[string]string{
	"[config]": ColorBlue,
	"[skip]":   ColorYellow,
	"[start]":  ColorGreen,
	"[error]":  ColorRed,
	"[reload]": ColorYellow,
	"[watch]":  ColorCyan,
}

func colorLog(format string, args ...interface{}) {
	color := ""
	for tag, c := range tagColors {
		if strings.HasPrefix(format, tag) {
			color = c
			break
		}
	}
	if color != "" {
		log.Printf(color+format+ColorReset, args...)
	} else {
		log.Printf(format, args...)
	}
}

func main() {
	if autostart.IsAnotherInstanceRunning() {
		fmt.Println("Another instance is already running.")
		return
	}

	log.Printf("%ssimulator_autostart %s started%s", ColorCyan, autostart.VERSION, ColorReset)

	engine := autostart.NewEngine(colorLog)
	engine.LoadConfig()

	log.Printf("%sType 'reload' to reload config.%s", ColorCyan, ColorReset)

	engine.WatchConfigFile()

	// Goroutine to listen for manual reload
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(input) == "reload" {
				engine.Reload()
			}
		}
	}()

	for {
		engine.RunOnce()
		time.Sleep(5 * time.Second)
	}
}
