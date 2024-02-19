package web

import (
	"fmt"
	"panopticon/lib"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func (p *PanelServer) ProcessStatus(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
