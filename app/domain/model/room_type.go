package model

import (
	"gorm.io/gorm"
)

type RoomType struct {
	gorm.Model
	ID             string `gorm:"type:varchar(36);primaryKey"`
	Name           string
	MaxParticipant int
	Description    string
	Rooms          []Room
}
