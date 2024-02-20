package web

import (
	"bytes"
	"fmt"
	"net/http"
	"panopticon/lib"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func (p *PanelServer) LogsLive(c echo.Context) error {
	name := c.QueryParam("proc")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provide query param `proc`")
	}
	offsetStr := c.QueryParam("offset")
	offset, err := strconv.Atoi(offsetStr)
	if name == "" || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "provide query param `offset`")
	}

	runningProcesses, err := p.Runner.GetRunningProcesses()
	if err != nil {
		return fmt.Errorf("could not get running processes list")
	}

	var maybeRunningProcess *lib.RunningProcess
	for _, v := range runningProcesses {
		if v.Proc.Name == name {
			maybeRunningProcess = &v
			break
		}
	}
	if maybeRunningProcess == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("proc not yet run %s", name))
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		t, err := tail.TailFile(maybeRunningProcess.LogPath, tail.Config{
			Follow: true,
			Location: &tail.SeekInfo{
				Offset: int64(offset),
			},
		})

		for line := range t.Lines {
			x := bytes.NewBuffer([]byte{})
			c.Echo().Renderer.Render(x, "log_line", line.Text, c)
			websocket.Message.Send(ws, x.String())
		}

		if err != nil {
			c.Logger().Fatal(err)
			return
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
