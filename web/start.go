package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (p *PanelServer) Start(c echo.Context) error {
	name := c.QueryParam("proc")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provide query param `proc`")
	}

	availableProcesses, err := p.Runner.GetAvailableProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes list")
	}
	exists := false
	for _, v := range availableProcesses {
		if v.Name == name {
			exists = true
		}
	}
	if !exists {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("proc does not exist %s", name))
	}

	runningProcesses, err := p.Runner.GetRunningProcesses()
	if err != nil {
		return fmt.Errorf("could not get running processes list")
	}
	exists = false
	for _, v := range runningProcesses {
		if v.Proc.Name == name && !v.Finished {
			exists = true
		}
	}
	if exists {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("proc already running %s", name))
	}

	proc, err := p.Runner.StartProcess(name)
	if err != nil {
		return fmt.Errorf("could not start process %s: '%s'", name, err.Error())
	}

	return c.Render(http.StatusOK, "proc", ProcessViewModel{
		Name:    proc.Proc.Name,
		Running: !proc.Finished,
		HasLogs: true,
	})
}
