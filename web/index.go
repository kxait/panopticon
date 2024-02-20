package web

import (
	"fmt"
	"net/http"
	"panopticon/lib"

	"github.com/labstack/echo/v4"
)

type ProcessViewModel struct {
	Name    string
	Running bool
	HasLogs bool
}

func (p *PanelServer) Index(c echo.Context) error {
	availableProcesses, err := p.Runner.GetAvailableProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes!")
	}
	runningProcs, err := p.Runner.GetRunningProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes!")
	}

	procs := make([]ProcessViewModel, len(availableProcesses))

	for k, v := range availableProcesses {
		var maybeRunningProc *lib.RunningProcess = nil
		for _, vv := range runningProcs {
			if v.Name == vv.Proc.Name {
				maybeRunningProc = &vv
			}
		}

		procs[k] = ProcessViewModel{
			Name:    v.Name,
			Running: maybeRunningProc != nil && !maybeRunningProc.Finished,
			HasLogs: maybeRunningProc != nil,
		}
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"procs": procs,
	})
}
