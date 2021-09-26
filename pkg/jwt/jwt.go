// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var jwtSecret []byte

// JwtGen jwt auth
func JwtGen(username string, exp time.Duration) (t string, err error) {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username

	claims["exp"] = now.Add(exp * time.Second).Unix()
	t, err = token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("JWT Generate Failure")
	}
	return t, nil
}

// JwtParse jwt parse
func JwtParse(tokenstring string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("Token invalid")
	}
	claim := token.Claims.(jwt.MapClaims)
	return claim, nil
}
