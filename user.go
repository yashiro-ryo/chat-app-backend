package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/net/websocket"
)

// NOTE 構造体のフィールド名は大文字でないとエンコードされない
type Response struct {
	ResponseType string `json:"response"`
	Status       int    `json:"status"`
	Body         `json:"body"`
}

type ResultUser struct {
	UserId   int    `json:"userId"`
	UserName string `json:"userName"`
	Message  string `json:"message"`
}

type Body struct {
	ResultUser `json:"resultUser"`
}

type GetMyInfoResponse struct {
	ResponseType string `json:"response"`
	Status       int    `json:"status"`
	MyInfoBody   `json:"body"`
}

type MyInfoBody struct {
	UserId   int    `json:"userId"`
	UserName string `json:"userName"`
}

func HandleGetMyInfo(token string, ws *websocket.Conn) error {
	userId, userName, err := GetMyInfo(token)
	if err != nil {
		return err
	}

	if userId == 0 {
		fmt.Println("userId is not found")
		return errors.New("userId is not found")
	}
	response := GetMyInfoResponse{
		ResponseType: "get-myinfo-result",
		Status:       200,
		MyInfoBody: MyInfoBody{
			userId,
			userName,
		},
	}

	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	if err != nil {
		return err
	}
	return nil
}

func HandleGetUserInfo(email string, ws *websocket.Conn) error {
	userInfo, err := GetUserInfo(email)
	// エラーハンドリングは上位の層に任せる
	var response Response
	if userInfo.UserId != nil && userInfo.UserName != nil && err == nil {
		response = Response{
			ResponseType: "search-user-result",
			Status:       200,
			Body: Body{
				ResultUser: ResultUser{
					UserId:   *userInfo.UserId,
					UserName: *userInfo.UserName,
				},
			},
		}
	} else {
		response = Response{
			ResponseType: "search-user-result",
			Status:       400,
			Body: Body{
				ResultUser: ResultUser{
					Message: "user is not found",
				},
			},
		}
	}

	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	return err
}

type UserInfo struct {
	UserName string `json:"userName"`
	UserId   int    `json:"userId"`
}

type GetUsersResponse struct {
	ResponseType string `json:"response"`
	Status       int    `json:"status"`
	GetUserBody  `json:"body"`
}

type GetUserBody struct {
	UserInfos []BasicUserInfo
}

func HandleGetUser(userId int, ws *websocket.Conn) {
	userInfos, err := GetUsers(userId)
	if err != nil {
		return
	}

	response := GetUsersResponse{
		ResponseType: "get-users-result",
		Status:       200,
		GetUserBody: GetUserBody{
			UserInfos: userInfos,
		},
	}

	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	if err != nil {
		return
	}
}

func HandleDeleteUser(userId int, removeUserId int, ws *websocket.Conn) {
	RemoveFriend(userId, removeUserId)
	HandleGetUser(userId, ws)
}

func HandleAddUser(userId int, addUserId int, ws *websocket.Conn) {
	AddFriend(userId, addUserId)
	HandleGetUser(userId, ws)
}
