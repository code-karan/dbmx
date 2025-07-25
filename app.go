package main

import (
	"context"
	"dbmx/app"
	"fmt"
)

// App struct
type App struct {
	ctx  context.Context
	conn *app.Connections
}

// NewApp creates a new App application struct
func NewApp(conn *app.Connections) *App {
	return &App{conn: conn}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	err := a.conn.TerminateAllDatabaseConnections()
	if err != nil {
		fmt.Println("Error terminating database connections:", err)
	}
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
