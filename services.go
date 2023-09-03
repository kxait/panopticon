package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/exp/slices"
)

func (s *Services) GetActiveProcesses() []ActiveProcess {
	return s.ActiveProcesses
}

func (s *Services) GetRunningProcesses() []ActiveProcess {
	active := s.ActiveProcesses
	result := make([]ActiveProcess, 0)
	for _, v := range active {
		if !v.Finished {
			result = append(result, v)
		}
	}

	return result
}

func (s *Services) isRunning(p *PanopProc) bool {
	for _, proc := range s.ActiveProcesses {
		if proc.Proc.Name == p.Name && !proc.Finished {
			return true
		}
	}

	return false
}

func (s *Services) GetRunnableProcesses() []PanopProc {
	result := make([]PanopProc, 0)
	for _, proc := range s.cfg.Procs {
		if !s.isRunning(&proc) {
			result = append(result, proc)
		}
	}
	return result
}

func (s *Services) StartProcess(p *PanopProc) error {
	if s.isRunning(p) {
		return fmt.Errorf(`already running: %s`, p.Name)
	}

	cmd := exec.Command(p.Cmd, p.Args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logPath := fmt.Sprintf("/tmp/%s.log", Sha256(p.Name))
	// logFile, err := os.OpenFile(logPath, os.O_CREATE, 0777)
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	cmd.Start()

	new := ActiveProcess{
		Finished: false,
		Proc:     *p,
		Cmd:      cmd,
		LogPath:  logPath,
		LogFile:  logFile,
	}

	var exitErr error = nil
	go func() {
		err := cmd.Wait()
		// if we're here, it exited...

		for i, runningProc := range s.ActiveProcesses {
			if runningProc.Proc.Name == p.Name {
				s.ActiveProcesses[i].Finished = true
			}
		}

		exitErr = err
		new.LogFile.Close()
	}()

	for k, v := range s.GetActiveProcesses() {
		if v.Proc.Name == p.Name {
			s.ActiveProcesses = slices.Delete(s.ActiveProcesses, k, k+1)
		}
	}
	s.ActiveProcesses = append(s.ActiveProcesses, new)

	// if we exited after 50ms then it failed to start
	time.Sleep(50 * time.Millisecond)
	if exitErr != nil {
		return fmt.Errorf("process failed to start: %s => %s", p.Name, exitErr.Error())
	}

	return nil
}

// func (p *ActiveProcess) PipeToOsStdout() {
// 	writer := io.MultiWriter(p.LogFile, os.Stdout)
// 	p.Cmd.Stdout = writer
// 	p.Cmd.Stderr = writer
// }

// func (p *ActiveProcess) PipeToFile() {
// 	p.Cmd.Stdout = p.LogFile
// 	p.Cmd.Stderr = p.LogFile
// }

// func (p *ActiveProcess) GetLogs() error {
// 	writer := bufio.NewWriter()

// 	fi, err := os.Open(p.LogPath)
// 	if err != nil {
// 		return err
// 	}

// 	writer.ReadFrom(fi)

// 	return nil
// }

func (s *Services) StopProcess(p *PanopProc, sig os.Signal) error {
	if !s.isRunning(p) {
		return fmt.Errorf(`not running: %s`, p.Name)
	}

	var c ActiveProcess
	for _, proc := range s.ActiveProcesses {
		if proc.Proc.Name == p.Name {
			c = proc
		}
	}

	return c.Cmd.Process.Signal(sig)
}

func (s *Services) Massacre() error {
	running := s.GetRunningProcesses()
	errors := make([]error, 0)

	for _, proc := range running {
		if err := syscall.Kill(-proc.Cmd.Process.Pid, syscall.SIGKILL); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		return &ErrorGroup{
			errs: errors,
		}
	}

	return nil
}
