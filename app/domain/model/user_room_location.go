package model

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type UserGameLocation struct {
	gorm.Model
	UserID uint
	User   *User
	RoomID uint
	Room   *Room
	XAxis  int
	YAxis  int
	Conn   *websocket.Conn `gorm:"-"`
	Mutex  sync.Mutex      `gorm:"-"`
}

func NewUserGameLocationByConn(conn *websocket.Conn) *UserGameLocation {
	return &UserGameLocation{Conn: conn}
}

func (u *UserGameLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserID uint `json:"userID"`
		RoomID uint `json:"roomID"`
		XAxis  int  `json:"xAxis"`
		YAxis  int  `json:"yAxis"`
	}{
		UserID: u.UserID,
		RoomID: u.RoomID,
		XAxis:  u.XAxis,
		YAxis:  u.YAxis,
	})
}
