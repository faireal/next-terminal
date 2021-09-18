// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ergoapi/zlog"
	"github.com/spf13/viper"
	"time"
)

var jwtSecret []byte

func getdefaultexp() int64 {
	exp := viper.GetInt64("core.login.exp")
	if exp > 0 {
		return exp
	}
	return 86400 // 1d 86400s 60 * 60 * 24
}

// JwtGen jwt auth
func JwtGen(username string, role string) (t string, err error) {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username

	expSecond := getdefaultexp()
	zlog.Debug("load token default exp: %v", expSecond)

	claims["exp"] = now.Add(time.Duration(expSecond) * time.Second).Unix()
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
