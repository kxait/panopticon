package web

import (
	"bytes"
	"fmt"
	"panopticon/lib"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func getProcRow(b *lib.Bussin, name string) (ProcessViewModel, error) {
	availableProcesses, err := b.GetAvailableProcesses()
	if err != nil {
		return ProcessViewModel{}, fmt.Errorf("could not get available processes!")
	}
	runningProcs, err := b.GetRunningProcesses()
	if err != nil {
		return ProcessViewModel{}, fmt.Errorf("could not get available processes!")
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

	if maybeAvailableProc == nil {
		return ProcessViewModel{}, fmt.Errorf("could not find process %s", name)
	}

	return ProcessViewModel{
		Name:    maybeAvailableProc.Name,
		Running: maybeRunningProc != nil && !maybeRunningProc.Finished,
		HasLogs: maybeRunningProc != nil,
	}, nil
}

func (p *PanelServer) ProcessStatus(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		listener := p.Runner.ProcessStatusNotifier.Subscribe()
		defer p.Runner.ProcessStatusNotifier.Unsubscribe(listener)
		for {
			select {
			case notif := <-listener:
				{
					row, err := getProcRow(p.Runner, notif.Name)
					if err != nil {
						c.Logger().Error(err)
						break
					}

					x := bytes.NewBuffer([]byte{})
					c.Echo().Renderer.Render(x, "proc", row, c)
					err = websocket.Message.Send(ws, x.String())
					if err != nil {
						c.Logger().Error(err)
						break
					}
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
