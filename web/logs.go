package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"panopticon/lib"
	"strings"

	terminal "github.com/buildkite/terminal-to-html"
	"github.com/labstack/echo/v4"
)

type LogsViewModel struct {
	Name    string
	Lines   []template.HTML
	Size    int
	LogPath string
}

func (p *PanelServer) Logs(c echo.Context) error {
	name := c.QueryParam("proc")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provide query param `proc`")
	}
	availableProcesses, err := p.Runner.GetAvailableProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes!")
	}
	runningProcs, err := p.Runner.GetRunningProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes!")
	}

	var maybeAvailableProc *lib.Process
	var maybeRunningProc *lib.RunningProcess
	for _, v := range availableProcesses {
		if name == v.Name {
			maybeAvailableProc = &v
			break
		}
	}
	for _, vv := range runningProcs {
		if name == vv.Proc.Name {
			maybeRunningProc = &vv
			break
		}
	}

	if maybeAvailableProc == nil || maybeRunningProc == nil {
		return c.String(http.StatusBadRequest, "process was not yet run")
	}

	data, err := os.ReadFile(maybeRunningProc.LogPath)
	if err != nil {
		return fmt.Errorf("could not read log file %s: %s", maybeRunningProc.LogPath, err.Error())
	}

	// TODO: optimize
	lines := strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n")
	linesEscaped := make([]template.HTML, len(lines))
	for k, v := range lines {
		linesEscaped[k] = template.HTML(terminal.Render([]byte(v)))
	}

	return c.Render(http.StatusOK, "logs.html", LogsViewModel{
		Name:    name,
		Lines:   linesEscaped,
		Size:    len(data),
		LogPath: maybeRunningProc.LogPath,
	})
}
