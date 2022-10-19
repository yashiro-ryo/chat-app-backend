package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/net/websocket"
)

// talkroom の構造体
type Talkroom struct {
	TalkroomId   int    `json:"talkroomId"`
	TalkroomName string `json:"talkroomName"`
	IsOwner      bool   `json:"isOwner"`
}

// NOTE 構造体のフィールド名は大文字でないとエンコードされない
// TODO 構造体の名前をカテゴリごとに変更する
type TalkroomResponse struct {
	ResponseType string `json:"response"`
	Status       int    `json:"status"`
	TalkroomBody `json:"body"`
}

type TalkroomBody struct {
	Talkrooms []Talkroom `json:"talkrooms"`
}

// talkroom
func HandleGetTalkrooms(userId int, ws *websocket.Conn) error {
	if userId == 0 {
		return errors.New("invalid user id")
	}
	talkrooms, err := GetTalkrooms(userId)
	if err != nil {
		fmt.Println("error get talkrooms")
	}
	// response
	response := TalkroomResponse{
		ResponseType: "get-talkrooms-result",
		Status:       200,
		TalkroomBody: TalkroomBody{
			Talkrooms: talkrooms,
		},
	}
	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	if err != nil {
		fmt.Println("websocket failed json talkrooms")
	}
	return nil
}

// NOTE 構造体のフィールド名は大文字でないとエンコードされない
// TODO 構造体の名前をカテゴリごとに変更する
type MessageResponse struct {
	ResponseType string `json:"response"`
	Status       int    `json:"status"`
	MessageBody  `json:"body"`
}

type MessageBody struct {
	Message []Message `json:"message"`
}

// talk
func HandleGetMessage(talkroomId int, ws *websocket.Conn) error {
	// TODO メッセージを書き込む権限は設定しないとダメかも
	messages, err := GetTalkroomMessage(talkroomId)
	if err != nil {
		fmt.Println("error :", err)
	}

	// response
	response := MessageResponse{
		ResponseType: "get-message-result",
		Status:       200,
		MessageBody: MessageBody{
			Message: messages,
		},
	}
	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	if err != nil {
		fmt.Println("websocket failed json talkrooms")
	}
	return nil
}

func HandleAddMessage(userId int, talkroomId int, contentType string, content string, ws *websocket.Conn) error {
	// contentType によって処理を分ける
	// ファイル部分は未確定
	err := SetMessage(talkroomId, userId, content)
	if err != nil {
		return errors.New("failed to set message")
	}
	// TODO メッセージを書き込む権限は設定しないとダメかも
	messages, err := GetTalkroomMessage(talkroomId)
	if err != nil {
		fmt.Println("error :", err)
	}

	// response
	response := MessageResponse{
		ResponseType: "get-message-result",
		Status:       200,
		MessageBody: MessageBody{
			Message: messages,
		},
	}
	// 構造体をjsonにエンコードする
	encodedJson, _ := json.Marshal(response)

	fmt.Println("encoded json :", string(encodedJson))
	err = websocket.Message.Send(ws, string(encodedJson))
	if err != nil {
		fmt.Println("websocket failed json talkrooms")
	}
	return nil
}
