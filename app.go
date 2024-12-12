package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		return nil, errors.Join(errors.New("unable to connect to database"), dbErr)
	}

	err = TryInitDb(db)
	if err != nil {
		return nil, errors.Join(errors.New("unable to initialize database"), err)
	}

	signingKey := GetEnvRequired("JWT_SIGNING_KEY")
	keyBytes, err := base64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return nil, errors.Join(errors.New("unable to decode the signing key"), err)
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

func (app *App) SignalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	log.Printf("Error: %v", err.Error())
}

func (app *App) HandleGetToken(c *gin.Context) {
	var guid struct {
		Guid uuid.UUID `json:"guid"`
	}
	err := c.BindJSON(&guid)
	if err != nil {
		return
	}
	conn, err := app.db.Acquire(context.Background())
	if err != nil {
		app.SignalError(c, err)
		return
	}
	defer conn.Release()
	var userExists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT * FROM users WHERE guid = $1)", guid.Guid).Scan(&userExists)
	if err != nil {
		app.SignalError(c, err)
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"detail": "User does not exist"})
		return
	}

	refr := uuid.New()
	token, err := app.jwtContext.MakeToken(guid.Guid.String(), refr.String())
	if err != nil {
		app.SignalError(c, err)
		return
	}

	refreshToken := RefreshToken{RefreshUUID: refr, ClientIp: c.ClientIP()}
	hash, err := refreshToken.HashBcrypt()
	if err != nil {
		app.SignalError(c, err)
		return
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO refresh_tokens (refresh_id, hash) VALUES ($1, $2)", refr, hash)
	if err != nil {
		app.SignalError(c, err)
	}
	tokenPair := TokenPair{Access: token, Refresh: refreshToken}
	c.JSON(http.StatusOK, &tokenPair)
}

func (app *App) HandleRefreshToken(c *gin.Context) {
	// TODO
	c.Status(http.StatusNotImplemented)
}
