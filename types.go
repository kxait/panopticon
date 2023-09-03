package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PanopProc struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

type PanopConfig struct {
	Procs []PanopProc `yaml:"procs"`
}

// can either mean currently running or finished, this is so we can view logs historically
type ActiveProcess struct {
	Finished bool
	Proc     PanopProc
	Cmd      *exec.Cmd
	LogPath  string
	LogFile  *os.File
}

type MainMenuChoice string

var (
	MAINMENU_START_SERVICES MainMenuChoice = "Start services"
	MAINMENU_STOP_SERVICES  MainMenuChoice = "Stop services"
	MAINMENU_SHOW_LOGS      MainMenuChoice = "Show logs for a service"
	MAINMENU_SHOW_STATS     MainMenuChoice = "Show stats for all services"
	MAINMENU_QUIT           MainMenuChoice = "Quit"
)

type Services struct {
	cfg             *PanopConfig
	ActiveProcesses []ActiveProcess
}

type ErrorGroup struct {
	errs []error
}

func (e *ErrorGroup) Error() string {
	errorMessages := make([]string, len(e.errs))
	for i, v := range e.errs {
		errorMessages[i] = v.Error()
	}
	return fmt.Sprintf("This error group contains %d errors, here they are: %s", len(e.errs), strings.Join(errorMessages, ", "))
}
