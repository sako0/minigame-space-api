package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirebaseUID string `gorm:"type:varchar(255);uniqueIndex"`
	Username    string
	AvatarID    uint
}

func NewUser(firebaseUID string) *User {
	return &User{
		FirebaseUID: firebaseUID,
	}
}
