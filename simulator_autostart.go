package main

import (
	"fmt"
	"github.com/mitchellh/go-ps"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

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
			log.Printf("Skip: Process %s is already running.\n", program.GetExecutable())
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
		log.Printf("Started: Process %s started successfully.\n", program.Path)
	}
	return nil
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
	configFile := filepath.Join(configDir, "config.ini")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModePerm)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := "" +
			"[AMS2AVX.exe]\n" +
			"programs=C:\\Program Files (x86)\\Britton IT Ltd\\CrewChiefV4\\CrewChiefV4.exe," +
			"C:\\Program Files (x86)\\SimHub\\SimHubWPF.exe," +
			"C:\\Users\\mail\\AppData\\Local\\popometer\\popometer-recorder.exe," +
			"C:\\Users\\mail\\work\\sim-to-motec\\ams2-cli.bat\n" +
			"[iRacingUI.exe]\n" +
			"programs=C:\\Program Files (x86)\\Britton IT Ltd\\CrewChiefV4\\CrewChiefV4.exe," +
			"C:\\Program Files (x86)\\SimHub\\SimHubWPF.exe," +
			"C:\\Users\\mail\\AppData\\Roaming\\garage61-install\\garage61-launcher.exe," +
			"C:\\Users\\mail\\AppData\\Local\\racelabapps\\RacelabApps.exe\n"
		os.WriteFile(configFile, []byte(defaultConfig), 0644)
	}

	cfg, err := ini.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	log.Println("Parsed Config:")
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		log.Printf("Process: %s", section.Name())
		programs := section.Key("programs").Strings(",")
		for _, program := range programs {
			log.Printf("  - %s", program)
		}
	}

	var processConfigs []ProcessConfig
	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		var programs []Program
		for _, program := range section.Key("programs").Strings(",") {
			programs = append(programs, Program{Path: program})
		}
		processConfigs = append(processConfigs, ProcessConfig{
			ProcessName:     section.Name(),
			ProgramsToStart: programs,
		})
	}

	return processConfigs
}

func main() {
	log.Println("simulator_autostart started")
	s := &State{}
	processConfigs := readProcessConfigs()

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
		log.Printf("Autostarting processes for %s ...\n", processName)
		err := startProcessesIfNotRunning(programsToStart)
		if err == nil {
			s.autostartedProcessIds = append(s.autostartedProcessIds, processId)
		} else {
			log.Println("Error:", err)
		}
	}
}
