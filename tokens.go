package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenPair struct {
	Access       string       `json:"access"`
	AccessParsed *jwt.Token   `json:"-"`
	Refresh      RefreshToken `json:"refresh"`
}

type RefreshToken struct {
	RefreshUUID uuid.UUID
	ClientIp    string
}

func (token *RefreshToken) MarshalText() []byte {
	return []byte(strings.Join([]string{token.RefreshUUID.String(), token.ClientIp}, "|"))
}

func (token *RefreshToken) UnmarshalText(text []byte) error {
	split := strings.Split(string(text), "|")
	if len(split) != 2 {
		return errors.New("invalid refresh token")
	}
	refreshID, err := uuid.Parse(split[0])
	if err != nil {
		return errors.Join(errors.New("invalid refresh token"), err)
	}
	token.RefreshUUID = refreshID
	token.ClientIp = split[1]
	return nil
}

func (token *RefreshToken) toBase64() string {
	return base64.StdEncoding.EncodeToString(token.MarshalText())
}

func (token *RefreshToken) fromBase64(encoded string) error {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	err = token.UnmarshalText(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (token *RefreshToken) MarshalJSON() ([]byte, error) {
	ret, err := json.Marshal(token.toBase64())
	return ret, err
}

func (token *RefreshToken) UnmarshalJSON(obj []byte) error {
	var encoded string
	err := json.Unmarshal(obj, &encoded)
	if err != nil {
		return err
	}
	err = token.fromBase64(encoded)
	return err
}

func (token *RefreshToken) getBcryptPassword() [64]byte {
	return sha512.Sum512(token.MarshalText())
}

func (token *RefreshToken) HashBcrypt() ([]byte, error) {
	password := token.getBcryptPassword()
	return bcrypt.GenerateFromPassword(password[:], bcrypt.DefaultCost)
}

func (token *RefreshToken) ValidateHash(hash []byte) error {
	password := token.getBcryptPassword()
	return bcrypt.CompareHashAndPassword(hash, password[:])
}
