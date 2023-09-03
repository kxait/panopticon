package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

func bussy() {

	dateCmd := exec.Command("date")

	dateOut, err := dateCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("> date")
	fmt.Println(string(dateOut))

	_, err = exec.Command("date", "-x").Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing:", err)
		case *exec.ExitError:
			fmt.Println("command exit rc =", e.ExitCode())
		default:
			panic(err)
		}
	}

	grepCmd := exec.Command("grep", "hello")

	grepIn, _ := grepCmd.StdinPipe()
	grepOut, _ := grepCmd.StdoutPipe()
	grepCmd.Start()
	grepIn.Write([]byte("hello grep\ngoodbye grep"))
	grepIn.Close()
	grepBytes, _ := io.ReadAll(grepOut)
	grepCmd.Wait()

	fmt.Println("> grep hello")
	fmt.Println(string(grepBytes))

	lsCmd := exec.Command("bash", "-c", "ls -a -l -h")
	lsOut, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("> ls -a -l -h")
	fmt.Println(string(lsOut))
}

func getId[K interface{}](list []K, item K, comparer func(a K, b K) bool) int {
	for k, v := range list {
		if comparer(v, item) {
			return k
		}
	}
	return -1
}

func validateConfig(c *PanopConfig) error {
	procs := c.Procs

	for k, v := range procs {
		if getId(procs, v, func(a, b PanopProc) bool {
			return a.Name == b.Name
		}) != k {

			return fmt.Errorf("duplicate process: %s", v.Name)
		}
	}

	return nil
}

func main() {
	buf, err := os.ReadFile("panop.yaml")
	if err != nil {
		panic("file not found panop.yaml")
	}
	c := &PanopConfig{}
	err = yaml.Unmarshal(buf, c)

	if err != nil {
		panic("invalid yaml panop.yaml")
	}

	if err := validateConfig(c); err != nil {
		panic(err.Error())
	}

	services := Services{
		cfg:             c,
		ActiveProcesses: make([]ActiveProcess, 0),
	}

	var options []string
	options = append(options, string(MAINMENU_START_SERVICES))
	options = append(options, string(MAINMENU_STOP_SERVICES))
	options = append(options, string(MAINMENU_SHOW_LOGS))
	options = append(options, string(MAINMENU_SHOW_STATS))
	options = append(options, string(MAINMENU_QUIT))

	for {
		selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
		fmt.Println()

		switch selectedOption {
		case string(MAINMENU_START_SERVICES):
			startServices(&services)
		case string(MAINMENU_STOP_SERVICES):
			stopServices(&services)
		case string(MAINMENU_SHOW_LOGS):
			showLogs(&services)
		case string(MAINMENU_QUIT):
			if err := services.Massacre(); err != nil {
				pterm.Error.Printfln("could not kill all processes: %s", err.Error())
			}
			os.Exit(0)
		}
	}
}

func startServices(s *Services) {
	pterm.Info.Println("Select the services to start")
	runnable := s.GetRunnableProcesses()
	procNames := make([]string, len(runnable))
	for i, runnableProc := range runnable {
		procNames[i] = runnableProc.Name
	}

	selected, _ := pterm.DefaultInteractiveMultiselect.WithOptions(procNames).Show()

	selectedRunnables := make([]PanopProc, len(selected))
	for i, selectedName := range selected {
		for _, possibleRunnable := range runnable {
			if possibleRunnable.Name == selectedName {
				selectedRunnables[i] = possibleRunnable
			}
		}
	}

	for _, selectedRunnable := range selectedRunnables {
		err := s.StartProcess(&selectedRunnable)
		if err != nil {
			pterm.Error.Printfln(`Process failed to start: %s (%s)`, pterm.Red(selectedRunnable.Name), pterm.Red(err.Error()))
		} else {
			pterm.Info.Printfln(`Process started successfully: %s`, pterm.Green(selectedRunnable.Name))
		}
	}

}

func stopServices(s *Services) {

	pterm.Info.Println("Select the services to stop")
	runnable := s.GetRunningProcesses()
	procNames := make([]string, len(runnable))
	for i, runnableProc := range runnable {
		procNames[i] = runnableProc.Proc.Name
	}

	selected, _ := pterm.DefaultInteractiveMultiselect.WithOptions(procNames).Show()

	selectedRunnables := make([]ActiveProcess, len(selected))
	for i, selectedName := range selected {
		for _, possibleRunnable := range runnable {
			if possibleRunnable.Proc.Name == selectedName {
				selectedRunnables[i] = possibleRunnable
			}
		}
	}

	for _, selectedRunnable := range selectedRunnables {
		err := s.StopProcess(&selectedRunnable.Proc, os.Kill)
		if err != nil {
			pterm.Error.Printfln(`Process failed to stop: %s (%s)`, pterm.Red(selectedRunnable.Proc.Name), pterm.Red(err.Error()))
		} else {
			pterm.Info.Printfln(`Process stopped successfully: %s`, pterm.Green(selectedRunnable.Proc.Name))
		}
	}

}

func showLogs(s *Services) {
	active := s.GetActiveProcesses()
	lips := make([]string, len(active))
	for k, v := range active {
		status := (func() string {
			if v.Finished {
				return pterm.Red("finished")
			}
			return pterm.Green("running")
		})()
		lips[k] = fmt.Sprintf("%s (%s)", v.Proc.Name, status)
	}

	selected, _ := pterm.DefaultInteractiveSelect.WithOptions(lips).Show()

	var proc ActiveProcess
	for k, v := range lips {
		if v == selected {
			proc = active[k]
		}
	}

	cmd := exec.Command("tail", "-n", "+0", "-f", proc.LogPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}
