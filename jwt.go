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

func (ctx *JWTContext) ParseToken(token string) (*jwt.Token, error) {
	return ctx.parser.Parse(token, ctx.keyFunc)
}

func (ctx *JWTContext) keyFunc(token *jwt.Token) (interface{}, error) {
	return ctx.signingKey, nil
}

func (ctx *JWTContext) MakeToken(sub, refr string) (string, error) {
	token := jwt.NewWithClaims(ctx.signingMethod, jwt.MapClaims{"sub": sub, "refr": refr})
	return token.SignedString(ctx.signingKey)
}
