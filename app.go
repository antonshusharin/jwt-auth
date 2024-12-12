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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (app *App) generateTokenPair(conn *pgxpool.Conn, userGuid uuid.UUID, clientIP string) (*TokenPair, error) {
	refr := uuid.New()
	token, err := app.jwtContext.MakeToken(userGuid.String(), refr.String())
	if err != nil {
		return nil, err
	}

	refreshToken := RefreshToken{RefreshUUID: refr, ClientIp: clientIP}
	hash, err := refreshToken.HashBcrypt()
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO refresh_tokens (refresh_id, hash) VALUES ($1, $2)", refr, hash)
	if err != nil {
		return nil, err
	}
	tokenPair := TokenPair{Access: token, Refresh: refreshToken}
	return &tokenPair, nil
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

	tokenPair, err := app.generateTokenPair(conn, guid.Guid, c.ClientIP())
	if err != nil {
		app.SignalError(c, err)
		return
	}
	c.JSON(http.StatusOK, &tokenPair)
}

func (app *App) HandleRefreshToken(c *gin.Context) {
	var tokenPair TokenPair
	err := c.BindJSON(&tokenPair)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Bad request"})
		return
	}

	accessToken, err := app.jwtContext.ParseToken(tokenPair.Access)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
		return
	}
	claims := accessToken.Claims.(jwt.MapClaims)
	refr := claims["refr"]
	sub := claims["sub"].(string)
	if refr != tokenPair.Refresh.RefreshUUID.String() {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
		return
	}

	conn, err := app.db.Acquire(context.Background())
	if err != nil {
		app.SignalError(c, err)
		return
	}
	defer conn.Release()
	var hash []byte
	err = conn.QueryRow(context.Background(), "SELECT hash FROM refresh_tokens WHERE refresh_id = $1", refr).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
		return
	}
	if err != nil {
		app.SignalError(c, err)
		return
	}

	err = tokenPair.Refresh.ValidateHash(hash)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Invalid token"})
		return
	}

	tx, err := conn.Begin(context.Background())
	if err != nil {
		app.SignalError(c, err)
		return
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(context.Background(), "DELETE FROM refresh_tokens WHERE refresh_id = $1", refr)
	if err != nil {
		app.SignalError(c, err)
		return
	}

	newTokenPair, err := app.generateTokenPair(conn, uuid.MustParse(sub), c.ClientIP())
	if err != nil {
		app.SignalError(c, err)
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		app.SignalError(c, err)
	}

	c.JSON(http.StatusOK, newTokenPair)
}
