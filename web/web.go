package web

import (
	"context"
	"embed"
	"html/template"
	"io"
	"panopticon/lib"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed page/fragments
var fragments embed.FS

//go:embed page
var page embed.FS

//go:embed page/static
var static embed.FS

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
	//e.Use(middleware.Recover())

	e.StaticFS("/static", static)

	//templates, _ := template.ParseGlob("web/page/fragments/*.html")
	templates, _ := template.ParseFS(fragments)
	//templates, _ = templates.ParseGlob("web/page/*.html")
	templates, _ = templates.ParseFS(page)
	t := &Template{
		templates: templates,
	}

	e.Renderer = t

	e.GET("/", p.Index)
	e.POST("/start", p.Start)
	e.POST("/stop", p.Stop)
	e.GET("/process-status", p.ProcessStatus)
	e.GET("/logs", p.Logs)
	e.GET("/logs-live", p.LogsLive)

	ctx, done := context.WithCancel(context.Background())
	go func() { p.Runner.ProcessStatusNotifier.Serve(p.Runner.ProcessStatusNotifierSource, ctx) }()

	e.Logger.Fatal(e.Start(":8080"))
	done()
}
