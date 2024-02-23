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

	for _, v := range os.Environ() {
		cmd.Env = append(cmd.Env, v)
	}
	for _, v := range proc.Env {
		cmd.Env = append(cmd.Env, v)
	}

	cmd.Start()

	go func(name string) {
		b.ProcessStatusNotifierSource <- ProcessStatusNotification{
			Name:    name,
			Running: true,
		}
	}(name)

	running := RunningProcess{
		Proc:     proc,
		Cmd:      cmd,
		LogPath:  logPath,
		LogFile:  logFile,
		Finished: false,
	}

	var exitErr error
	go func(name string, running RunningProcess) {
		err := cmd.Wait()
		// if we're here, it exited...

		for i, runningProc := range b.RunningProcesses {
			if runningProc.Proc.Name == name {
				b.RunningProcesses[i].Finished = true
				break
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
	}(proc.Name, running)

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

func (b *Bussin) KillAllChildren(sig syscall.Signal) []error {
	errs := make([]error, 0)
	var maybeRunningProcess *RunningProcess

	for _, v := range b.RunningProcesses {
		if !b.isRunning(v.Proc.Name) {
			continue
		}

		pgid, err := syscall.Getpgid(v.Cmd.Process.Pid)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not get PGID of process %s PID %d reason: %s", v.Proc.Name, maybeRunningProcess.Cmd.Process.Pid, err.Error()))
		}

		err = syscall.Kill(-pgid, sig)
		if err != nil {
			errs = append(errs, err)
		}

		time.Sleep(100 * time.Millisecond)
		if b.isRunning(v.Proc.Name) {
			errs = append(errs, fmt.Errorf("process still lived 100ms after killing: %s", v.Proc.Name))
		}
	}

	return errs
}

func (b *Bussin) KillProcess(name string, sig syscall.Signal) error {
	if !b.isRunning(name) {
		return fmt.Errorf("not running: %s", name)
	}

	var maybeRunningProcess *RunningProcess
	for _, v := range b.RunningProcesses {
		if v.Proc.Name == name {
			maybeRunningProcess = &v
			break
		}
	}
	if maybeRunningProcess == nil {
		return fmt.Errorf("could not find process %s", name)
	}

	pgid, err := syscall.Getpgid(maybeRunningProcess.Cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("could not get PGID of process %s PID %d reason: %s", name, maybeRunningProcess.Cmd.Process.Pid, err.Error())
	}

	return syscall.Kill(-pgid, sig)
}

func (b *Bussin) isRunning(name string) bool {
	for _, v := range b.RunningProcesses {
		if v.Proc.Name == name {
			return !v.Finished
		}
	}

	return false
}
