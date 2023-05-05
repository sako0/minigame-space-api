package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type UserRepository interface {
	GetUser(userId uint) (*model.User, error)
	AddUser(user *model.User) error
	RemoveUser(userId uint) error
	GetUserByFirebaseUID(firebaseUID string) (*model.User, error)
}
