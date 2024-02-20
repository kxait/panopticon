package lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/exp/slices"
)

type ProcessStatusNotification struct {
	Name    string
	Running bool
}

type Bussin struct {
	AvailableProcesses          []Process
	RunningProcesses            []RunningProcess
	ProcessStatusNotifier       Broadcaster[ProcessStatusNotification]
	ProcessStatusNotifierSource chan ProcessStatusNotification
}

type Process struct {
	Name string
	Cmd  string
	Cwd  string
	Args []string
	Env  []string
}

func (b *Bussin) GetAvailableProcesses() ([]Process, error) {
	if b.AvailableProcesses == nil {
		return nil, fmt.Errorf("nil available processes")
	}

	return b.AvailableProcesses, nil
}

type RunningProcess struct {
	Proc     Process
	Finished bool
	Cmd      *exec.Cmd
	LogPath  string
	LogFile  *os.File
}

func (b *Bussin) GetRunningProcesses() ([]RunningProcess, error) {
	if b.RunningProcesses == nil {
		return nil, fmt.Errorf("nil running processes")
	}

	return b.RunningProcesses, nil
}

func (b *Bussin) StartProcess(name string) (RunningProcess, error) {
	if b.isRunning(name) {
		return RunningProcess{}, fmt.Errorf("already running: %s", name)
	}

	var proc Process
	for _, v := range b.AvailableProcesses {
		if v.Name == name {
			proc = v
		}
	}

	if proc.Name != name {
		return RunningProcess{}, fmt.Errorf("process does not exist: %s", name)
	}

	for _, v := range proc.Env {
		result := strings.Split(v, "=")
		if len(result) == 1 {
			return RunningProcess{}, fmt.Errorf("invalid env entry: %s", v)
		}
	}

	cmd := exec.Command(proc.Cmd, proc.Args...)
	// set group ID so child dies together with parent
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logPath := fmt.Sprintf("/tmp/%s.log", Sha256(proc.Name))
	logFile, err := os.Create(logPath)

	if err != nil {
		return RunningProcess{}, err
	}

	if proc.Cwd != "" {
		cmd.Dir = proc.Cwd
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	cmd.Env = proc.Env

	cmd.Start()

	go func() {
		b.ProcessStatusNotifierSource <- ProcessStatusNotification{
			Name:    name,
			Running: true,
		}
	}()

	running := RunningProcess{
		Proc:     proc,
		Cmd:      cmd,
		LogPath:  logPath,
		LogFile:  logFile,
		Finished: false,
	}

	var exitErr error
	go func() {
		err := cmd.Wait()
		// if we're here, it exited...

		for i, runningProc := range b.RunningProcesses {
			if runningProc.Proc.Name == proc.Name {
				b.RunningProcesses[i].Finished = true
			}
		}
		go func() {
			b.ProcessStatusNotifierSource <- ProcessStatusNotification{
				Name:    name,
				Running: false,
			}
		}()

		exitErr = err
		running.LogFile.Close()
	}()

	// remove the previous (finished) process from the proc list so we can add the new one
	for k, v := range b.RunningProcesses {
		if v.Proc.Name == proc.Name {
			b.RunningProcesses = slices.Delete(b.RunningProcesses, k, k+1)
		}
	}
	b.RunningProcesses = append(b.RunningProcesses, running)

	// if we exited after 50ms then it failed to start
	time.Sleep(50 * time.Millisecond)
	if exitErr != nil {
		return RunningProcess{}, fmt.Errorf("process failed to start: %s => %s", proc.Name, exitErr.Error())
	}

	return running, nil
}

func (b *Bussin) KillProcess(name string, sig os.Signal) error {
	if !b.isRunning(name) {
		return fmt.Errorf("not running: %s", name)
	}

	var c RunningProcess
	for _, v := range b.RunningProcesses {
		if v.Proc.Name == name {
			c = v
		}
	}

	return c.Cmd.Process.Signal(sig)
}

// TODO: signature - some kind of pipe?
//func GetProcessLogs(name string) error {
//	return nil
//}

// TODO: signature - some new struct
//func GetProcessDetails(name string) error {
//	return nil
//}

func (b *Bussin) isRunning(name string) bool {
	for _, v := range b.RunningProcesses {
		if v.Proc.Name == name {
			return !v.Finished
		}
	}

	return false
}
