package model

import (
	"gorm.io/gorm"
)

type Area struct {
	gorm.Model
	ID             string `gorm:"type:varchar(36);primaryKey"`
	Name           string
	MaxParticipant int
	RoomCount      int
	Status         string
	Description    string
	Rooms          []Room
	UserLocations  []UserLocation
}
