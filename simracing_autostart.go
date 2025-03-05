package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
)

var autostartedProcessIds []int

func getProcessIdForExecutable(processName string) int {
	processes, _ := ps.Processes()

	for _, p := range processes {
		if p.Executable() == processName {
			return p.Pid()
		}
	}

	return -1
}

func startProcesses(exePaths ...string) error {
	for _, exePath := range exePaths {
		cmd := exec.Command("cmd.exe", "/C", "start", "", exePath)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start process %s: %w", exePath, err)
		}
		log.Printf("Process %s started successfully.\n", exePath)
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

func main() {
	log.Println("simracing_autostart started")
	for {
		processName := "AMS2AVX.exe"
		processId := getProcessIdForExecutable(processName)
		if processId != -1 && !contains(autostartedProcessIds, processId) {
			log.Printf("Autostarting processes for %s ...\n", processName)
			err := startProcesses(
				"C:\\Program Files (x86)\\Britton IT Ltd\\CrewChiefV4\\CrewChiefV4.exe",
				"C:\\Program Files (x86)\\SimHub\\SimHubWPF.exe",
				"C:\\Users\\mail\\AppData\\Local\\popometer\\popometer-recorder.exe",
				"C:\\Users\\mail\\work\\sim-to-motec\\ams2-cli.bat",
			)
			if err == nil {
				autostartedProcessIds = append(autostartedProcessIds, processId)
			} else {
				log.Println("Error:", err)
			}
		}
		time.Sleep(5 * time.Second)
	}
}
