package autostart

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

func ConfigFilePath() string {
	return filepath.Join(os.Getenv("APPDATA"), "simulator_autostart", "config.yaml")
}

func (e *Engine) readProcessConfigs() []ProcessConfig {
	configDir := filepath.Join(os.Getenv("APPDATA"), "simulator_autostart")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModePerm)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := `# Assetto Corsa
AssettoCorsa.exe:
  programs:
    - C:\Program Files (x86)\Britton IT Ltd\CrewChiefV4\CrewChiefV4.exe
    - C:\Program Files (x86)\SimHub\SimHubWPF.exe
    - C:\Users\mail\Telemetry_for_RaceSims\telemetry_tool\runWin_AC.bat
    # Example with custom working directory (use | to separate):
    # - C:\path\to\app.exe|D:\custom\workdir

# Flight Simulator 2024
FlightSimulator2024.exe:
  programs:
    - C:\Program Files (x86)\SimHub\SimHubWPF.exe
    - C:\Program Files\Little Navmap\littlenavmap.exe
`
		os.WriteFile(configFile, []byte(defaultConfig), 0644)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		e.logFn("[error] Failed to read config file: %v", err)
		return nil
	}

	var config map[string]struct {
		Programs []string `yaml:"programs"`
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		e.logFn("[error] Failed to parse config file: %v", err)
		return nil
	}

	e.logFn("[config] Parsed Config from: %s", configFile)
	var processConfigs []ProcessConfig

	var processNames []string
	for processName := range config {
		processNames = append(processNames, processName)
	}
	sort.Strings(processNames)

	for _, processName := range processNames {
		processData := config[processName]
		e.logFn("[config] Process: %s", processName)
		var programs []Program
		for _, programPath := range processData.Programs {
			program := parseProgram(programPath)
			if program.WorkDir != "" {
				e.logFn("[config]   - %s (workdir: %s)", program.Path, program.WorkDir)
			} else {
				e.logFn("[config]   - %s", program.Path)
			}
			programs = append(programs, program)
		}
		processConfigs = append(processConfigs, ProcessConfig{
			ProcessName:     processName,
			ProgramsToStart: programs,
		})
	}

	return processConfigs
}

func (e *Engine) WatchConfigFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		e.logFn("[error] Failed to create file watcher: %v", err)
		return
	}

	configDir := filepath.Dir(ConfigFilePath())
	err = watcher.Add(configDir)
	if err != nil {
		e.logFn("[error] Failed to watch config directory: %v", err)
		watcher.Close()
		return
	}

	e.logFn("[watch] Watching config file for changes.")

	go func() {
		defer watcher.Close()
		var debounceTimer *time.Timer
		configName := filepath.Base(ConfigFilePath())

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) != configName {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
					e.Reload()
				})
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				e.logFn("[error] File watcher error: %v", err)
			}
		}
	}()
}

func parseProgram(input string) Program {
	parts := strings.Split(input, "|")
	program := Program{Path: strings.TrimSpace(parts[0])}
	if len(parts) > 1 {
		program.WorkDir = strings.TrimSpace(parts[1])
	}
	return program
}
