package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          string `gorm:"type:varchar(36);primaryKey"`
	FirebaseUID string `gorm:"type:varchar(255);uniqueIndex"`
	Username    string
	AvatarID    string `gorm:"type:varchar(36)"`
}

func NewUser(firebaseUID string) *User {
	return &User{
		FirebaseUID: firebaseUID,
	}
}
