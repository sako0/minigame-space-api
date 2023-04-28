package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type UserRepository interface {
	GetUser(firebaseUid string) (*model.User, error)
	AddUser(user *model.User) error
	UpdateUser(user *model.User) error
}
