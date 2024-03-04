package main

import (
	"fmt"
	"os"
	"os/signal"
	"panopticon/lib"
	"panopticon/web"
	"syscall"
)

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		path = "panop.yaml"
	}

	rawConfig, err := lib.ReadConfig(path)
	if err != nil {
		panic(fmt.Sprintf("could not read config %s (%s)", path, err.Error()))
	}

	config := lib.MapConfig(rawConfig)
	fmt.Printf("%+v\n", config)

	server := web.PanelServer{
		Runner: config,
	}

	// IF I DIE ALL THE KIDS DIE TOO
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)

	go func() {
		for {
			sig := <-sigchnl
			switch sig {

			case syscall.SIGCHLD:
			case syscall.SIGURG:
			case syscall.SIGPIPE:
			case syscall.SIGWINCH:
			case syscall.SIGPROF:
				break
			default:
				{
					fmt.Printf("%+v\n", sig)
					errs := config.KillAllChildren(syscall.SIGTERM)
					for _, v := range errs {
						fmt.Printf("error reaping process: %s\n", v.Error())
					}
					fmt.Println("goodbye!")
					os.Exit(0)
				}
			}
		}
	}()

	server.Serve()
}

//func showLogs(s *Services) {
//	active := s.GetActiveProcesses()
//	lips := make([]string, len(active))
//	for k, v := range active {
//		status := (func() string {
//			if v.Finished {
//				return pterm.Red("finished")
//			}
//			return pterm.Green("running")
//		})()
//		lips[k] = fmt.Sprintf("%s (%s)", v.Proc.Name, status)
//	}
//
//	empty := false
//
//	selected, err := pterm.DefaultInteractiveSelect.WithOptions(lips).WithOnInterruptFunc(func() {
//		empty = true
//	}).Show()
//
//	// time.Sleep(10 * time.Millisecond)
//
//	if err != nil || empty {
//		return
//	}
//
//	var proc ActiveProcess
//	for k, v := range lips {
//		if v == selected {
//			proc = active[k]
//		}
//	}
//
//	cmd := exec.Command("tail", "-n", "+0", "-f", proc.LogPath)
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	cmd.Stdin = os.Stdin
//	cmd.Run()
//}
