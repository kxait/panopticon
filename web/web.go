package web

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"panopticon/lib"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type PanelServer struct {
	Runner *lib.Bussin
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (p *PanelServer) Serve() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &Template{
		templates: template.Must(template.ParseGlob("web/page/*.html")),
	}

	e.Renderer = t

	e.GET("/", p.index)

	e.Logger.Fatal(e.Start(":8080"))
}

func (p *PanelServer) index(c echo.Context) error {
	availableProcesses, err := p.Runner.GetAvailableProcesses()
	if err != nil {
		return fmt.Errorf("could not get available processes!")
	}

	processNames := make([]string, len(availableProcesses))
	for k, v := range availableProcesses {
		processNames[k] = v.Name
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"procs": processNames,
	})
}
