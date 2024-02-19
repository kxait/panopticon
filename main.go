package main

import (
	"fmt"
	"os"
	"panopticon/lib"
	"panopticon/web"
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
