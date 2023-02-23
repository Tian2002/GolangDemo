package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JwtUserClaims struct {
	ID int64
	jwt.RegisteredClaims
}

//var stSigningKey []byte = []byte(config.Config.JWT.StSigningKey)
var stSigningKey []byte = []byte("key")

func GenerateToken(id int64) (string, error) {
	iJwtUserClaims := JwtUserClaims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3000 * time.Minute)), //持续时间
			//ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.JWT.ExpiresTime) * time.Minute)), //持续时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  "Token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, iJwtUserClaims)
	return token.SignedString(stSigningKey)
}

func ParseToken(tokenStr string) (int64, error) {
	iJwtUserClaims := JwtUserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &iJwtUserClaims, func(token *jwt.Token) (interface{}, error) {
		return stSigningKey, nil
	})
	if err == nil && !token.Valid {
		err = errors.New("invalid token") //无效的token
	}

	return iJwtUserClaims.ID, err
}
