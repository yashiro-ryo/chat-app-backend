// database の単体テスト
package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchUserInfo(t *testing.T) {
	// 正常データ
	var id int = 1
	var name string = "矢代 涼"
	var email string = "aaaa@aaaa.com"
	SetupDB()
	userInfo, err := GetUserInfo(email)

	if err != nil {
		t.Errorf("failed NewUser()")
	}

	if userInfo.UserId == nil || userInfo.UserName == nil {
		t.Errorf("user info is null")
	} else {
		assert.Equal(t, id, userInfo.UserId)
		assert.Equal(t, name, userInfo.UserName)
	}

	CloseDB()
}
