package web

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func (p *PanelServer) ProcessStatus(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		listener := p.Runner.ProcessStatusNotifier.Subscribe()
		defer p.Runner.ProcessStatusNotifier.Unsubscribe(listener)
		for {
			select {
			case notif := <-listener:
				{
					msg, err := json.Marshal(notif)
					if err != nil {
						c.Logger().Error(err)
						break
					}
					err = websocket.Message.Send(ws, string(msg))
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
