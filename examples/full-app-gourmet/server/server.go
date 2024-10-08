package server

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/examples/full-app-gourmet/controller"
	"github.com/go-fuego/fuego/examples/full-app-gourmet/static"
	"github.com/go-fuego/fuego/examples/full-app-gourmet/templates"
	"github.com/go-fuego/fuego/examples/full-app-gourmet/views"
)

type Resources struct {
	Views views.Resource
	API   controller.Resource
}

func (rs Resources) Setup(
	options ...func(*fuego.Server),
) *fuego.Server {
	serverOptions := []func(*fuego.Server){
		fuego.WithAutoAuth(controller.LoginFunc),
		fuego.WithTemplateFS(templates.FS),
		fuego.WithTemplateGlobs("**/*.html", "**/**/*.html"),
		fuego.WithGlobalResponseTypes(http.StatusForbidden, "Forbidden"),
	}

	options = append(serverOptions, options...)

	// Create server with some options
	app := fuego.NewServer(options...)

	app.OpenApiSpec.Info.Title = "Gourmet API"

	rs.API.Security = app.Security

	// Register middlewares (functions that will be executed before AND after the controllers, in the order they are registered)
	// With fuego, you can use any existing middleware that relies on `net/http`, or create your own
	fuego.Use(app, chiMiddleware.Compress(5, "text/html", "text/css", "application/json"))

	fuego.Register(app, fuego.Route[any, any]{
		Path:     "/static/",
		Method:   http.MethodGet,
		FullName: "static handler",
	}, http.StripPrefix("/static", static.Handler()), func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=600")

			h.ServeHTTP(w, r)
		})
	})

	// Register views (controllers that return HTML pages)
	rs.Views.Routes(fuego.Group(app, "/"))

	// Register API routes (controllers that return JSON)
	rs.API.MountRoutes(fuego.Group(app, "/api"))

	return app
}
