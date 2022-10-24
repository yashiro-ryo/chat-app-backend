package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 取得したデータをいれる構造体を準備する
type Person struct {
	user_id  int
	password string
}

var db *sql.DB
var err error

/*
func main() {
	SetupDB()
	fmt.Println(CheckLogin("aaaa@aaaa", "aiueo"))
	CloseDB()
}
*/

// system
func SetupDB() {
	db, err = sql.Open("mysql", "root:LTDEXPuzushio22@@tcp(localhost:3306)/chat-app?parseTime=true")
	if err != nil {
		// ここではエラーを返さない
		log.Fatal(err)
	}
}

func CloseDB() {
	defer db.Close()
}

// login singin auth user
// NOTE メールアドレスとパスワードを認証する関数 成功時はuser_id 失敗時はnil

func CheckLogin(email string, pass string) int {
	// db にインスタンスが渡っていなければreturn
	if db == nil {
		return 0
	}
	// SQLの実行
	rows, err := db.Query("SELECT user_id, pass_hashed FROM user WHERE email = '" + email + "'")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	// TODO 不要な抽出データの削減(構造体の変更+spl文の変更)
	var counter = 0
	var userId = 0
	for rows.Next() {
		counter++
		var person Person
		err := rows.Scan(&person.user_id, &person.password)
		if err != nil {
			panic(err.Error())
		}
		log.Println("test :", person.user_id, person.password)
		userId = person.user_id
	}
	if counter != 0 {
		return userId
	} else {
		return 0
	}
}

// token と reflesh tokenをdbに登録する(認証サーバーのような感じ)
func SaveToken(userId int, token string, tokenExpiredAt int64, refleshToken string, refleshTokenExpiredAt int64) error {
	if db == nil {
		return errors.New("db is not found")
	}
	rows, err := db.Query("SELECT user_id FROM token WHERE user_id = " + strconv.Itoa(userId))
	if err != nil {
		log.Fatal(err)
	}
	var counter int
	for rows.Next() {
		counter++
	}
	if counter == 1 {
		// データがある場合はupdate
		ins, err := db.Prepare("UPDATE token SET user_id=?, token=?, token_expired_at=?, reflesh_token=?, reflesh_token_expired_at=? WHERE user_id = " + strconv.Itoa(userId))
		if err != nil {
			log.Fatal(err)
		}
		ins.Exec(userId, token, tokenExpiredAt, refleshToken, refleshTokenExpiredAt)
	} else {
		// データがない場合は挿入(migrate)
		ins, err := db.Prepare("INSERT INTO token (user_id, token, token_expired_at, reflesh_token, reflesh_token_expired_at) VALUES(?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		ins.Exec(userId, token, tokenExpiredAt, refleshToken, refleshTokenExpiredAt)
	}
	return nil
}

// talk
func GetTalkrooms(userId int) ([]Talkroom, error) {
	if db == nil {
		return nil, errors.New("db is not found")
	}
	//var strUserId = string(userId)
	rows, err := db.Query("select talkroom_id from talkroom_user where user_id = 1")
	if err != nil {
		log.Fatal(err)
	}
	var talkroomResult []Talkroom
	for rows.Next() {
		var talkroomId int
		if err := rows.Scan(&talkroomId); err != nil {
			log.Fatal(err)
		}

		rowss, err := db.Query("select talkroom_name, is_deleted from talkroom where talkroom_id = " + strconv.Itoa(talkroomId))
		if err != nil {
			log.Fatal(err)
		}

		for rowss.Next() {
			var talkroomName string
			var isDeleted bool
			if err = rowss.Scan(&talkroomName, &isDeleted); err != nil {
				log.Fatal(err)
			}

			talkroom := Talkroom{
				TalkroomId:   talkroomId,
				TalkroomName: talkroomName,
				IsOwner:      isDeleted,
			}
			fmt.Println(talkroom)
			talkroomResult = append(talkroomResult, talkroom)
		}
	}
	return talkroomResult, nil
}

func SetMessage(talkroomId int, sentUserId int, message string) error {
	if db == nil {
		return errors.New("db is not found")
	}

	var idStr = `message_` + strconv.Itoa(talkroomId)
	fmt.Println(idStr)

	// insert
	ins, err := db.Prepare("INSERT INTO " + idStr + " (message_id, content_type, content, sent_user_id, sent_at) VALUES(?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	ins.Exec(nil, "message", message, sentUserId, "2022/09/12 09:33:21")
	return nil
}

type Message struct {
	Message_id   int
	Content_type string
	Content      string
	Sent_user_id int
	Sent_at      string
}

func GetTalkroomMessage(talkroomId int) ([]Message, error) {
	if db == nil {
		return nil, errors.New("db is not found")
	}

	var idStr = `message_` + strconv.Itoa(talkroomId)

	rows, err := db.Query("select message_id, content_type, content, sent_user_id, sent_at from " + idStr)
	if err != nil {
		log.Fatal(err)
	}
	var userResult []Message
	for rows.Next() {
		var id int
		var types string
		var con string
		var sentUserId int
		var at string
		if err := rows.Scan(&id, &types, &con, &sentUserId, &at); err != nil {
			log.Fatal(err)
		}
		user := Message{
			Message_id:   id,
			Content_type: types,
			Content:      con,
			Sent_user_id: sentUserId,
			Sent_at:      at,
		}
		userResult = append(userResult, user)
	}
	return userResult, nil
}

// address
// userIdとuserNameの構造体
type BasicUserInfo struct {
	UserId   *int    `json:"userId"`
	UserName *string `json:"userName"`
}

func GetUserInfo(email string) (BasicUserInfo, error) {
	var userInfo BasicUserInfo
	// db にインスタンスが渡っていなければreturn
	if db == nil {
		return userInfo, errors.New("db is not found")
	}
	// SQLの実行
	rows, err := db.Query("SELECT user_id, name FROM user WHERE email = '" + email + "'")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	// 1データのみを受け付ける(想定は1データ以上あるとやばい)
	for rows.Next() {
		err = rows.Scan(&userInfo.UserId, &userInfo.UserName)
		fmt.Println("serach user :", *userInfo.UserId, *userInfo.UserName)
	}
	return userInfo, err
}

func AddFriend(userId int, addUserId int) error {
	if db == nil {
		return errors.New("db is not found")
	}
	// insert
	ins, err := db.Prepare("INSERT INTO user_friend (user_id, friend_user_id) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	ins.Exec(userId, addUserId)
	return nil
}

func RemoveFriend(user_id int, removeUserId int) error {
	if db == nil {
		return errors.New("db is not found")
	}
	// insert
	ins, err := db.Prepare("DELETE FROM user_friend WHERE user_id = ? AND friend_user_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	ins.Exec(user_id, removeUserId)
	return nil
}

func GetUsers(userId int) ([]BasicUserInfo, error) {
	var userInfos []BasicUserInfo
	// db にインスタンスが渡っていなければreturn
	if db == nil {
		return nil, errors.New("db is not found")
	}
	// SQLの実行
	rows, err := db.Query("SELECT friend_user_id FROM user_friend WHERE user_id = " + strconv.Itoa(userId))
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var userId int
		var userName string
		err = rows.Scan(&userId)
		rowss, err := db.Query("SELECT user_id, name FROM user WHERE user_id = " + strconv.Itoa(userId))
		if err != nil {
			return nil, errors.New("query error")
		}
		// １件のみのヒットでなければやばい
		for rowss.Next() {
			rowss.Scan(&userId, &userName)
		}
		fmt.Println("serach user :", userId, userName)
		user := BasicUserInfo{
			UserId:   &userId,
			UserName: &userName,
		}
		userInfos = append(userInfos, user)
	}
	return userInfos, err
}

func CreateTalkroomData(ownerUser int, talkroomName string) (int, error) {
	if db == nil {
		return 0, errors.New("db is not found")
	}
	// insert
	ins, err := db.Prepare("INSERT INTO talkroom (talkroom_id, talkroom_name, is_deleted) VALUES(?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	rand.Seed(time.Now().UnixNano())
	var randInt = rand.Intn(1000-1) + 1
	ins.Exec(nil, talkroomName, randInt)
	rows, err := db.Query("SELECT talkroom_id FROM talkroom WHERE is_deleted = " + strconv.Itoa(randInt))
	if err != nil {
		log.Fatal(err)
	}
	var talkroomId = 0
	for rows.Next() {
		rows.Scan(&talkroomId)
	}
	return talkroomId, nil
}

func BindUserIdForCreateTalkroom(bindUserId int, talkroomId int) error {
	if db == nil {
		return errors.New("db is not found")
	}
	// insert
	ins, err := db.Prepare("INSERT INTO talkroom_user (user_id, talkroom_id, is_owner) VALUES(?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	ins.Exec(bindUserId, talkroomId, 1)
	fmt.Println("insert into talkroom user")
	return nil
}

func CreateTalkroomTable(talkroomId int) error {
	if db == nil {
		return errors.New("db is not found")
	}
	// insert
	_, err := db.Query("CREATE TABLE IF NOT EXISTS message_" + strconv.Itoa(talkroomId) + " (message_id int auto_increment not null primary key, content_type varchar(255) not null, content varchar(10) not null, sent_user_id int not null, sent_at datetime)")
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func ResetDeleteStatus(talkroomId int) error {
	if db == nil {
		return errors.New("db is not found")
	}
	// insert
	_, err := db.Query("UPDATE talkroom SET is_deleted = 0 WHERE talkroom_id = " + strconv.Itoa(talkroomId))
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// 取る必要にある関数
func GetMyInfo(token string) (int, string, error) {
	if db == nil {
		return 0, "", errors.New("db is not found")
	}

	rows, err := db.Query("SELECT user_id, name FROM user WHERE user_id = (SELECT user_id FROM token WHERE token = '" + token + "')")
	if err != nil {
		return 0, "", errors.New("query error")
	}
	var userId int
	var userName string
	for rows.Next() {
		rows.Scan(&userId, &userName)
	}
	return userId, userName, nil
}
