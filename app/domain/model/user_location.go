package model

import (
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type UserLocation struct {
	gorm.Model
	UserID   uint `gorm:"unique"`
	User     User
	AreaID   uint
	Area     Area
	RoomID   uint
	Room     Room
	XAxis    int
	YAxis    int
	JoinedAt int
	Conn     *websocket.Conn `gorm:"-"`
}

func NewUserLocation(user *User, room *Room, area *Area, conn *websocket.Conn) *UserLocation {
	return &UserLocation{
		User: *user,
		Room: *room,
		Area: *area,
		Conn: conn,
	}
}
