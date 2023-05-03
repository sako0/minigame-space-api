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

func NewUserLocation(user *User, area *Area, room *Room, xAxis int, yAxis int, joinedAt int, conn *websocket.Conn) *UserLocation {
	return &UserLocation{
		UserID:   user.ID,
		User:     user,
		AreaID:   area.ID,
		Area:     area,
		RoomID:   room.ID,
		Room:     room,
		XAxis:    xAxis,
		YAxis:    yAxis,
		JoinedAt: joinedAt,
		Conn:     conn,
	}
}
