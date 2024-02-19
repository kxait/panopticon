package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func (p *PanelServer) Stop(c echo.Context) error {
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
		if v.Proc.Name == name {
			exists = true
		}
	}
	if !exists {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("proc not running %s", name))
	}

	err = p.Runner.KillProcess(name, os.Kill)
	if err != nil {
		return fmt.Errorf("could not start process %s: '%s'", name, err.Error())
	}

	return c.Render(http.StatusOK, "proc", ProcessViewModel{
		Name:    name,
		Running: false,
		HasLogs: true,
	})
}
