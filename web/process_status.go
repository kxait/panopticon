package web

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func (p *PanelServer) ProcessStatus(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			notif := <-p.Runner.ProcessStatusNotifier
			msg, err := json.Marshal(notif)
			if err != nil {
				c.Logger().Error(err)
				break
			}
			err = websocket.Message.Send(ws, string(msg))
			if err != nil {
				c.Logger().Error(err)
				p.Runner.ProcessStatusNotifier <- notif
				break
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
