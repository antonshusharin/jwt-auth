package main

import "github.com/golang-jwt/jwt/v5"

type JWTContext struct {
	signingKey    []byte
	parser        *jwt.Parser
	signingMethod jwt.SigningMethod
}

func NewJWTContext(signingKey []byte) *JWTContext {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS512"}))
	signingMethod := jwt.SigningMethodHS512
	return &JWTContext{signingKey, parser, signingMethod}
}

func (ctx *JWTContext) ParseToken(token string) *jwt.Token {
	// TODO
	return new(jwt.Token)
}

func (ctx *JWTContext) MakeToken(sub, refr string) string {
	// TODO
	return ""
}
