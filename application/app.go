package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	route http.Handler
}

func New() *App {
	app := &App{
		route: loadRoutes(),
	}

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.route,
	}
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}
	return nil
}
