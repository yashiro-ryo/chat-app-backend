package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

func GenerateToken(tokenType string, userId int) (string, error) {
	var expiredTime int64
	// reflesh token の場合は有効期限を少し長くする
	if tokenType == "token" {
		expiredTime = time.Now().Add(time.Hour * 24).Unix()
	} else {
		expiredTime = time.Now().Add(time.Hour * 36).Unix()
	}
	// ペイロードの作成
	claims := jwt.MapClaims{
		"token-type": tokenType,
		"user_id":    12345678,
		"exp":        expiredTime,
	}

	// トークン生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// トークンに署名を付与
	tokenString, err := token.SignedString([]byte("SECRET_KEY"))
	if err != nil {
		return "", err
	}
	return tokenString, err
}
