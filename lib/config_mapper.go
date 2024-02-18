package lib

func MapConfig(c *Config) *Bussin {
	procs := make([]Process, len(c.Procs))
	runningProcesses := make([]RunningProcess, 0)

	for k, v := range c.Procs {
		procs[k] = Process{
			Name: v.Name,
			Cmd:  v.Cmd,
			Args: v.Args,
			Env:  c.Env,
		}
	}

	return &Bussin{
		AvailableProcesses: procs,
		RunningProcesses:   runningProcesses,
	}
}
