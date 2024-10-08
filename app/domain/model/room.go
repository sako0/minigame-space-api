package model

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	AreaID        uint
	Area          Area
	RoomTypeID    uint
	RoomType      RoomType
	Status        int
	UserLocations []UserLocation
}
