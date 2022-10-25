package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ログイン処理＆トークン生成
func login(c echo.Context) error {
	email := c.FormValue("email")
	pass := c.FormValue("password")

	fmt.Println("form data :", email, pass)

	var userId int = CheckLogin(email, pass)
	fmt.Println("user id :", userId)
	// username, passwordの確認
	if userId == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"errorMsg": "メールアドレスまたはパスワードが不正です",
		})
	}

	token, tokenExpiredAt, err := GenerateToken("token", userId)
	if err != nil {
		return err
	}
	refleshToken, refleshTokenExpiredAt, err := GenerateToken("refleshtoken", userId)
	if err != nil {
		return err
	}

	// dbにtokenとreflesh tokenを保存する
	err = SaveToken(userId, token, tokenExpiredAt, refleshToken, refleshTokenExpiredAt)
	if err != nil {
		return errors.New("failed save token to db")
	}

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = token
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	cookie.Name = "refleshToken"
	cookie.Value = refleshToken
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	fmt.Println(token, refleshToken)

	return c.JSON(http.StatusOK, echo.Map{
		"token":        token,
		"refleshtoken": refleshToken,
	})
}

func CheckToken(tokenString string) (interface{}, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("SECRET_KEY"), nil
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

// ユーザ情報取得
func user(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := int64(claims["user_id"].(float64))
	return c.String(http.StatusOK, fmt.Sprintf("userID: %v", userID))
}

// 認証ルーティング
func SetupJwtRouter(e echo.Echo) {
	// ログイン処理(token認証不要)
	e.POST("/login", login)

	// user group(token認証必要)
	r := e.Group("/user")

	// echo.middleware JWTConfigの設定
	config := middleware.JWTConfig{
		SigningKey: []byte("SECRET_KEY"),
		ParseTokenFunc: func(tokenString string, c echo.Context) (interface{}, error) {
			keyFunc := func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte("SECRET_KEY"), nil
			}

			token, err := jwt.Parse(tokenString, keyFunc)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("", user)
}
