package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mitchellh/go-ps"
)

const VERSION = "0.4"

type Program struct {
	Path    string
	WorkDir string
}

func (p Program) GetExecutable() string {
	return filepath.Base(p.Path)
}

func (p Program) GetFolder() string {
	if p.WorkDir != "" {
		return p.WorkDir
	}
	return filepath.Dir(p.Path)
}

type ProcessConfig struct {
	ProcessName     string
	ProgramsToStart []Program
}

type Engine struct {
	processConfigs     []ProcessConfig
	startedProcesses   []Program
	autostartedProcIDs []int
	logFn              func(string, ...interface{})
}

func NewEngine(logFn func(string, ...interface{})) *Engine {
	return &Engine{logFn: logFn}
}

func (e *Engine) LoadConfig() {
	e.processConfigs = e.readProcessConfigs()
}

func (e *Engine) Reload() {
	e.logFn("[reload] Soft reload requested...")
	e.startedProcesses = nil
	e.autostartedProcIDs = nil
	e.processConfigs = e.readProcessConfigs()
}

func (e *Engine) RunOnce() {
	for _, cfg := range e.processConfigs {
		e.startProgramsIfProcessIsRunning(cfg.ProcessName, cfg.ProgramsToStart)
	}
}

func (e *Engine) startProgramsIfProcessIsRunning(processName string, programsToStart []Program) {
	processID := GetProcessIDForExecutable(processName)
	if processID != -1 && !containsInt(e.autostartedProcIDs, processID) {
		e.logFn("[start] Autostarting processes for %s ...", processName)
		err := e.startProcessesIfNotRunning(programsToStart)
		if err == nil {
			e.autostartedProcIDs = append(e.autostartedProcIDs, processID)
		} else {
			e.logFn("[error] %v", err)
		}
	}
}

func (e *Engine) startProcessesIfNotRunning(programs []Program) error {
	for _, program := range programs {
		if GetProcessIDForExecutable(program.GetExecutable()) != -1 {
			e.logFn("[skip] Process %s is already running.", program.GetExecutable())
			continue
		}

		if isNonExeTool(program) && e.hasBeenStartedBefore(program) {
			e.logFn("[skip] Non exe process %s has been started before.", program.GetExecutable())
			continue
		}

		cmd := exec.Command("cmd.exe", "/C", "start", "", program.Path)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		cmd.Dir = program.GetFolder()

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start process %s: %w", program.Path, err)
		}
		if program.WorkDir != "" {
			e.logFn("[start] Process %s started successfully (workdir: %s).", program.Path, program.WorkDir)
		} else {
			e.logFn("[start] Process %s started successfully.", program.Path)
		}

		e.startedProcesses = append(e.startedProcesses, program)
	}
	return nil
}

func (e *Engine) hasBeenStartedBefore(program Program) bool {
	for _, p := range e.startedProcesses {
		if p.Path == program.Path {
			return true
		}
	}
	return false
}

// IsAnotherInstanceRunning checks if another process with the same executable
// name is already running (excluding our own PID).
func IsAnotherInstanceRunning() bool {
	self := filepath.Base(os.Args[0])
	myPID := os.Getpid()
	processes, _ := ps.Processes()
	for _, p := range processes {
		if p.Executable() == self && p.Pid() != myPID {
			return true
		}
	}
	return false
}

func GetProcessIDForExecutable(processName string) int {
	processes, _ := ps.Processes()
	for _, p := range processes {
		if p.Executable() == processName {
			return p.Pid()
		}
	}
	return -1
}

func isNonExeTool(program Program) bool {
	return !strings.HasSuffix(program.Path, ".exe")
}

func containsInt(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
