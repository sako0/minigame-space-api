package model

import (
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
}

func NewUserLocation(user *User, room *Room, area *Area) *UserLocation {
	return &UserLocation{
		User: *user,
		Room: *room,
		Area: *area,
	}
}
