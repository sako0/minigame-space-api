package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type UserRepository interface {
	GetUser(userId string) (*model.User, error)
	GetUserByFirebaseUID(firebaseUID string) (*model.User, error)
}
