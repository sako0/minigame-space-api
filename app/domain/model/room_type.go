package model

import (
	"gorm.io/gorm"
)

type RoomType struct {
	gorm.Model
	Name           string
	MaxParticipant int
	Description    string
	Rooms          []Room
}
