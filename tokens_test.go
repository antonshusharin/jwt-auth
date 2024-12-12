package main

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenRoundtrip(t *testing.T) {
	assert := assert.New(t)
	token := RefreshToken{RefreshUUID: uuid.New(), ClientIp: "127.0.0.1"}
	encoded, err := json.Marshal(&token)
	assert.NoError(err, "Could not encode token")
	var token2 RefreshToken
	err = json.Unmarshal(encoded, &token2)
	assert.NoError(err, "Could not decode token")
	assert.Equal(token, token2, "Encoded and decoded tokens do not match")
}

func TestRefreshTokensecure(t *testing.T) {
	assert := assert.New(t)
	uid := uuid.New()
	uid2 := uuid.New()
	token := RefreshToken{RefreshUUID: uid, ClientIp: "127.0.0.1"}
	tokenHash, err := token.HashBcrypt()
	assert.NoError(err, "Unable to hash token")
	err = token.ValidateHash(tokenHash)
	assert.NoError(err, "Token did not validate against its own hash")
	badToken := token
	badToken.RefreshUUID = uid2
	badTokenHash, err := badToken.HashBcrypt()
	assert.NoError(err, "Unable to hash token")
	err = token.ValidateHash(badTokenHash)
	assert.Error(err, "Validated a bad token hash")
}
