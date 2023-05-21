package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type UserGameLocationRepository interface {
	GetUserGameLocation(userId uint) (*model.UserGameLocation, bool, error)
	AddUserGameLocation(userLocation *model.UserGameLocation) error
	RemoveUserGameLocation(userId uint) error
	UpdateUserGameLocation(userGameLocation *model.UserGameLocation) error
	GetAllUserGameLocationsByRoomId(roomId uint) ([]*model.UserGameLocation, bool, error)
}
