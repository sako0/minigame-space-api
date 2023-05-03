package model

import (
	"gorm.io/gorm"
)

type Area struct {
	gorm.Model
	Name           string
	MaxParticipant int
	RoomCount      int
	Status         string
	Description    string
	Rooms          []Room
	UserLocations  []UserLocation
}
