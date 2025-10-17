package main

import (
	"bufio"
	"fmt"
	"github.com/mitchellh/go-ps"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

const VERSION = "0.4"

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

var startedProcesses []Program
var processConfigs []ProcessConfig // moved global so reload can update it

func getProcessIdForExecutable(processName string) int {
	processes, _ := ps.Processes()

	for _, p := range processes {
		if p.Executable() == processName {
			return p.Pid()
		}
	}

	return -1
}

func startProcessesIfNotRunning(programs []Program) error {
	for _, program := range programs {
		if getProcessIdForExecutable(program.GetExecutable()) != -1 {
			log.Printf("%sSkip:%s Process %s is already running.\n", ColorYellow, ColorReset, program.GetExecutable())
			continue
		}

		if isNonExeTool(program) && hasBeenStartedBefore(program) {
			log.Printf("%sSkip:%s Non exe process %s has been started before.\n", ColorYellow, ColorReset, program.GetExecutable())
			continue
		}

		cmd := exec.Command("cmd.exe", "/C", "start", "", program.Path)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		cmd.Dir = program.GetFolder() // Set working directory

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start process %s: %w", program.Path, err)
		}
		log.Printf("%sStarted:%s Process %s started successfully.\n", ColorGreen, ColorReset, program.Path)

		startedProcesses = append(startedProcesses, program)
	}
	return nil
}

func hasBeenStartedBefore(program Program) bool {
	for _, process := range startedProcesses {
		if process.Path == program.Path {
			return true
		}
	}
	return false
}

func isNonExeTool(program Program) bool {
	if !strings.HasSuffix(program.Path, ".exe") {
		return true
	}
	return false
}

func contains(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

type State struct {
	autostartedProcessIds []int
}

type Program struct {
	Path string
}

func (p Program) GetExecutable() string {
	return filepath.Base(p.Path)
}

func (p Program) GetFolder() string {
	return filepath.Dir(p.Path)
}

type ProcessConfig struct {
	ProcessName     string
	ProgramsToStart []Program
}

func readProcessConfigs() []ProcessConfig {
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
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]struct {
		Programs []string `yaml:"programs"`
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	log.Printf("%sParsed Config from:%s %s\n", ColorBlue, ColorReset, configFile)
	var processConfigs []ProcessConfig

	// Get sorted process names
	var processNames []string
	for processName := range config {
		processNames = append(processNames, processName)
	}
	sort.Strings(processNames)

	// Process in alphabetical order
	for _, processName := range processNames {
		processData := config[processName]
		log.Printf("%sProcess:%s %s", ColorPurple, ColorReset, processName)
		var programs []Program
		for _, programPath := range processData.Programs {
			log.Printf("  - %s", programPath)
			programs = append(programs, Program{Path: programPath})
		}
		processConfigs = append(processConfigs, ProcessConfig{
			ProcessName:     processName,
			ProgramsToStart: programs,
		})
	}

	return processConfigs
}

func main() {
	log.Printf("%ssimulator_autostart %s started%s\n", ColorCyan, VERSION, ColorReset)

	// Initial config load
	processConfigs = readProcessConfigs()
	s := &State{}
	log.Printf("%sType 'reload' to reload config.%s\n", ColorCyan, ColorReset)

	// Goroutine to listen for "r" key
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(input) == "reload" {
				log.Printf("%sSoft reload requested...%s\n", ColorYellow, ColorReset)
				startedProcesses = nil
				s = &State{} // reset state
				processConfigs = readProcessConfigs()
				log.Printf("%sType 'reload' to reload config.%s\n", ColorCyan, ColorReset)
			}
		}
	}()

	for {
		for _, config := range processConfigs {
			s.startProgramsIfProcessIsRunning(config.ProcessName, config.ProgramsToStart)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *State) startProgramsIfProcessIsRunning(processName string, programsToStart []Program) {
	processId := getProcessIdForExecutable(processName)
	if processId != -1 && !contains(s.autostartedProcessIds, processId) {
		log.Printf("%sAutostarting processes for %s ...%s\n", ColorGreen, processName, ColorReset)
		err := startProcessesIfNotRunning(programsToStart)
		if err == nil {
			s.autostartedProcessIds = append(s.autostartedProcessIds, processId)
		} else {
			log.Printf("%sError:%s %v\n", ColorRed, ColorReset, err)
		}
	}
}
