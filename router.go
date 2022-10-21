package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

// リクエストの構造体を定義
type RequestStruct struct {
	Token   string `json:"token"`
	Request string `json:"request"`
	// data は中身がリクエストごとに異なるのでハンドリング先でバリデーションをかけてやる
	Data struct {
		UserId       *int    `json:"userId"`
		UserName     *string `json:"userName"`
		QueryEmail   *string `json:"queryEmail"`
		TalkroomName *string `json:"talkroomName"`
		UserIds      *[]int  `json:"userIds"`
		TalkroomId   *int    `json:"talkroomId"`
		ContentType  *string `json:"contentType"`
		Content      *string `json:"content"`
		DeleteUserId *int    `json:"deleteUserId"`
		AddUserId    *int    `json:"addUserId"`
	}
}

func handleWebSocket(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// 初回のメッセージを送信
		err := websocket.Message.Send(ws, ``)
		if err != nil {
			c.Logger().Error(err)
		}

		for {
			// Client からのメッセージを読み込む
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
			}

			// clientからきた文字列jsonを変数として使えるようにデコードする
			var request RequestStruct
			err := json.Unmarshal([]byte(msg), &request)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("request :", request.Request)

			// token取得
			cookie, err := c.Cookie("token")
			if err != nil {
				fmt.Println("cookie error", err)
				return
			}
			fmt.Println("cookie name", cookie.Name)
			fmt.Println("cookie value", cookie.Value)
			// 有効期限切れの場合は再度ログイン(reflesh token はまだ実装しない)
			token, err := CheckToken(cookie.Value)
			if err != nil {
				fmt.Println(err)
				fmt.Println("failed to check token")
				c.Redirect(http.StatusOK, "http://localhost:8080/signin")
				return
			}
			user := token.(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			exp := int64(claims["exp"].(float64))
			fmt.Println("token expired at", exp)
			userId := float64(claims["user_id"].(float64))
			fmt.Println("token-re", userId)

			switch request.Request {
			case "get-myinfo":
				fmt.Println("called get myinfo")
				HandleGetMyInfo(cookie.Value, ws)
			case "search-user":
				fmt.Println("serach-user")
				if request.Data.QueryEmail != nil {
					HandleGetUserInfo(*request.Data.QueryEmail, ws)
				} else {
					fmt.Println("not found email")
				}
			case "get-user":
				fmt.Println("get-user")
				if request.Data.UserId != nil {
					HandleGetUser(*request.Data.UserId, ws)
				} else {
					fmt.Println("not found user id")
				}
			case "add-user":
				fmt.Println("add user")
				if request.Data.UserId != nil && request.Data.AddUserId != nil {
					HandleAddUser(*request.Data.UserId, *request.Data.AddUserId, ws)
				} else {
					fmt.Println("failed to add user")
				}
			case "delete-user":
				fmt.Println("delete-user")
				if request.Data.UserId != nil && request.Data.DeleteUserId != nil {
					HandleDeleteUser(*request.Data.UserId, *request.Data.DeleteUserId, ws)
				} else {
					fmt.Println("failed to delete user")
				}
			case "get-talkrooms":
				fmt.Println("get-talkrooms")
				if request.Data.UserId != nil {
					HandleGetTalkrooms(*request.Data.UserId, ws)
				}
			case "create-talkroom":
				fmt.Println("create-talkroom")
				if request.Data.UserId != nil && request.Data.UserIds != nil && request.Data.TalkroomName != nil {
					HandleCreateTalkroom(*request.Data.UserId, *request.Data.UserIds, *request.Data.TalkroomName, ws)
				} else {
					fmt.Println("fail to create talkroom")
				}
			case "delete-talkroom":
			case "get-message":
				fmt.Println("get-message")
				if request.Data.TalkroomId != nil {
					HandleGetMessage(*request.Data.TalkroomId, ws)
				} else {
					fmt.Println("not found contentType or content")
				}
			case "add-message":
				fmt.Println("add-message")
				if request.Data.UserId != nil && request.Data.TalkroomId != nil && request.Data.ContentType != nil && request.Data.Content != nil {
					err := HandleAddMessage(*request.Data.UserId, *request.Data.TalkroomId, *request.Data.ContentType, *request.Data.Content, ws)
					if err != nil {
						fmt.Println(err)
					}
				}
			}

			/*
				// token 認証
				// talkroomにメッセージを追加
				err = SetMessage(int(map1["talkroomId"].(float64)), int(map1["sentUserId"].(float64)), map1["content"].(string))
				if err != nil {
					c.String(http.StatusOK, `{"error":"db error"}`)
				}
				var userResult []Message
				// talkroomからメッセージを取得
				userResult, err = GetTalkroomMessage(int(map1["talkroomId"].(float64)))
				if err != nil {
					c.String(http.StatusOK, `{"error":"db error"}`)
				}

				if err != nil {
					c.Logger().Error(err)
				}

				data1, _ := json.Marshal(userResult)
			*/

			// Client からのメッセージを元に返すメッセージを作成し送信する
			err = websocket.Message.Send(ws, `{"message": "banana"}`)
			if err != nil {
				c.Logger().Error(err)
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func SetupRouter() {
	e := echo.New()
	e.Use(middleware.Logger())
	SetupJwtRouter(*e)

	// file
	e.Static("/", "public")
	e.File("/signin", "public/signin.html")
	e.File("/signup", "public/signup.html")
	e.GET("/recovery", func(c echo.Context) error {
		return c.String(http.StatusOK, "リカバリーページ")
	})

	// routing
	e.GET("/ws", handleWebSocket)

	e.Logger.Fatal(e.Start(":8080"))
}
