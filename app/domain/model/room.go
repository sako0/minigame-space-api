package model

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	ID            string `gorm:"type:varchar(36);primaryKey"`
	AreaID        string `gorm:"type:varchar(36)"`
	Area          Area
	RoomTypeID    string `gorm:"type:varchar(36)"`
	RoomType      RoomType
	Status        int
	UserLocations []UserLocation
}

func NewRoom(areaId, roomTypeId string) *Room {
	return &Room{AreaID: areaId, RoomTypeID: roomTypeId, Status: 0}
}
