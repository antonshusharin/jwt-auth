package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	router     *gin.Engine
	db         *pgxpool.Pool
	jwtContext *JWTContext
}

func AppFromEnvironment() *App {
	// TODO
	app := App{router: gin.Default(), db: nil, jwtContext: NewJWTContext([]byte{})}
	app.registerRoutes()
	app.router.HandleMethodNotAllowed = true
	return &app
}

func (app *App) registerRoutes() {
	app.router.POST("/get-token", app.HandleGetToken)
	app.router.POST("/refresh-token", app.HandleRefreshToken)
}

func (app *App) Run() {
	app.router.Run("0.0.0.0:5000")
}

func (app *App) HandleGetToken(c *gin.Context) {
	// TODO
	c.Status(http.StatusNotImplemented)
}

func (app *App) HandleRefreshToken(c *gin.Context) {
	// TODO
	c.Status(http.StatusNotImplemented)
}
