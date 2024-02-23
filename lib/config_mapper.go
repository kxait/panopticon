package lib

import (
	"fmt"
)

func MapConfig(c *Config) *Bussin {
	procs := make([]Process, len(c.Procs))
	runningProcesses := make([]RunningProcess, 0)
	env := make([]string, 0)
	for k, v := range c.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range c.Procs {
		procs[k] = Process{
			Name: v.Name,
			Cmd:  v.Cmd,
			Cwd:  v.Cwd,
			Args: v.Args,
			Env:  env,
		}
	}

	sauce := make(chan ProcessStatusNotification)
	return &Bussin{
		AvailableProcesses:          procs,
		RunningProcesses:            runningProcesses,
		ProcessStatusNotifierSource: sauce,
		ProcessStatusNotifier:       Broadcaster[ProcessStatusNotification]{},
	}
}
