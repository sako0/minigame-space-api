package model

import (
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type UserLocation struct {
	gorm.Model
	UserID   uint
	User     *User
	AreaID   uint
	Area     *Area
	RoomID   uint
	Room     *Room
	XAxis    int
	YAxis    int
	JoinedAt int
	Conn     *websocket.Conn `gorm:"-"`
}

func NewUserLocationByConn(conn *websocket.Conn) *UserLocation {
	return &UserLocation{Conn: conn}
}
