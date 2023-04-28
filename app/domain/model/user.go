package model

import (
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirebaseUID string `gorm:"type:varchar(255);uniqueIndex"`
	Username    string
	AvatarID    uint
	Conn        *websocket.Conn `gorm:"-"`
}

func NewUser(conn *websocket.Conn, firebaseUID string) *User {
	return &User{
		FirebaseUID: firebaseUID,
		Conn:        conn,
	}
}
