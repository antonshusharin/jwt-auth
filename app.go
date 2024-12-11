package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	router     *gin.Engine
	db         *pgxpool.Pool
	jwtContext *JWTContext
}

func AppFromEnvironment() (*App, error) {
	dbHost := GetEnvOptional("JWT_DB_HOST", "localhost")
	dbPort := GetEnvOptional("JWT_DB_PORT", "5432")
	dbUser := GetEnvRequired("JWT_DB_USER")
	dbPassword := GetEnvOptional("JWT_DB_PASSWORD", "")
	dbDatabase := GetEnvOptional("JWT_DB_DATABASE", "")

	var connStringBuilder strings.Builder
	fmt.Fprintf(&connStringBuilder, "host=%v", dbHost)
	fmt.Fprintf(&connStringBuilder, " port=%v", dbPort)
	fmt.Fprintf(&connStringBuilder, " user=%v", dbUser)
	if dbPassword != "" {
		fmt.Fprintf(&connStringBuilder, " password=%v", dbPassword)
	}
	if dbDatabase != "" {
		fmt.Fprintf(&connStringBuilder, " dbname=%v", dbDatabase)
	}

	db, err := pgxpool.New(context.Background(), connStringBuilder.String())
	if err != nil {
		return nil, err
	}
	dbErr := db.Ping(context.Background())
	if dbErr != nil {
		return nil, dbErr
	}

	signingKey := GetEnvRequired("JWT_SIGNING_KEY")
	keyBytes, err := base64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return nil, errors.Join(errors.New("Unable to decode the signing key"), err)
	}

	app := App{router: gin.Default(), db: db, jwtContext: NewJWTContext(keyBytes)}
	app.registerRoutes()
	app.router.HandleMethodNotAllowed = true
	return &app, nil
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
