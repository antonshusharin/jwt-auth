package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var app *App

func TestMain(m *testing.M) {
	var err error
	app, err = AppFromEnvironment()
	if err != nil {
		log.Fatalf("Unable to initialize testing environment: %v", err.Error())
	}
	err = populateMockData(app.db)
	if err != nil {
		log.Fatalf("Unable to populate the database with mock data: %v", err.Error())
	}
	code := m.Run()
	destroyTestData(app.db)
	os.Exit(code)
}

func populateMockData(db *pgxpool.Pool) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	for _, user := range MOCK_USERS {
		_, err = tx.Exec(context.Background(), "INSERT INTO users (guid, username, email) values ($1, $2, $3)", user.guid, user.username, user.email)
		if err != nil {
			return err
		}
	}
	err = tx.Commit(context.Background())
	return err
}

func destroyTestData(db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "DROP TABLE users; DROP TABLE refresh_tokens")
	if err != nil {
		log.Printf("Notice: destroyTestData failed, future test runs might not succeede. %v", err.Error())
	}
}

func makeTestRequest(req *http.Request) (statusCode int, responseBody []byte) {
	w := httptest.NewRecorder()
	req.Header["Content-Type"] = []string{"application/json"}
	app.router.ServeHTTP(w, req)
	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, respBody
}

func TestGetAndRefreshToken(t *testing.T) {
	assert := assert.New(t)
	body := strings.NewReader(fmt.Sprintf(`{"guid": "%v"}`, MOCK_USERS["Ivan"].guid))
	req := httptest.NewRequest("POST", "/get-token", body)
	assert.NotNil(req)
	status, respBody := makeTestRequest(req)
	assert.Equal(http.StatusOK, status, "Could not retrieve access token")
	var tokenPair TokenPair
	err := json.Unmarshal(respBody, &tokenPair)
	assert.NoError(err, "Could not parse token pair as JSON")
	body2, _ := json.Marshal(tokenPair)
	req = httptest.NewRequest("POST", "/refresh-token", bytes.NewReader(body2))
	status, respBody = makeTestRequest(req)
	assert.Equal(http.StatusOK, status, "Could not refresh with provided token pair")
	err = json.Unmarshal(respBody, &tokenPair)
	assert.NoError(err, "Could not decode refreshed token pair")
	req = httptest.NewRequest("POST", "/refresh-token", bytes.NewReader(body2))
	status, _ = makeTestRequest(req)
	assert.Equal(http.StatusForbidden, status, "Was able to refresh the same token pair twice")
}

func TestJWTSecurity(t *testing.T) {
	assert := assert.New(t)
	conn, err := app.db.Acquire(context.Background())
	require.NoError(t, err, "Could not connect to database")
	defer conn.Release()
	pair1, err := app.generateTokenPair(conn, MOCK_USERS["Ivan"].guid, "192.0.2.1")
	require.NoError(t, err, "Could not create token pair 1")
	pair2, err := app.generateTokenPair(conn, MOCK_USERS["Maria"].guid, "192.0.2.1")
	require.NoError(t, err, "Could not create token pair 2")
	badPair := TokenPair{pair1.Access, pair2.Refresh}
	body, _ := json.Marshal(badPair)
	req := httptest.NewRequest("POST", "/refresh-token", bytes.NewReader(body))
	status, _ := makeTestRequest(req)
	assert.Equal(http.StatusForbidden, status, "Was able to refresh a token with that from another user!")

	body, _ = json.Marshal(pair2)
	req = httptest.NewRequest("POST", "/refresh-token", bytes.NewReader(body))
	req.RemoteAddr = "192.0.2.2"
	status, _ = makeTestRequest(req)
	assert.Equal(http.StatusOK, status, "Could not refresh pair 2, which is valid")
	mailer := app.mailer.(*TestingMailer)
	mail := mailer.CheckEmail()
	require.NotNil(t, mailer, "Warning email was not sent")
	assert.Equal([]string{MOCK_USERS["Maria"].email}, mail.To, "Warning email was sent to the wrong user")
}
