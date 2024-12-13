package main

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenRoundtrip(t *testing.T) {
	require := require.New(t)
	token := RefreshToken{RefreshUUID: uuid.New(), ClientIp: "127.0.0.1"}
	encoded, err := json.Marshal(&token)
	require.NoError(err, "Could not encode token")
	var token2 RefreshToken
	err = json.Unmarshal(encoded, &token2)
	require.NoError(err, "Could not decode token")
	require.Equal(token, token2, "Encoded and decoded tokens do not match")
}

func TestRefreshTokensecure(t *testing.T) {
	require := require.New(t)
	uid := uuid.New()
	uid2 := uuid.New()
	token := RefreshToken{RefreshUUID: uid, ClientIp: "127.0.0.1"}
	tokenHash, err := token.HashBcrypt()
	require.NoError(err, "Unable to hash token")
	err = token.ValidateHash(tokenHash)
	require.NoError(err, "Token did not validate against its own hash")
	badToken := token
	badToken.RefreshUUID = uid2
	badTokenHash, err := badToken.HashBcrypt()
	require.NoError(err, "Unable to hash token")
	err = token.ValidateHash(badTokenHash)
	require.Error(err, "Validated a bad token hash")
}
