package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type InMemoryUserGameLocationRepository interface {
	Store(userGameLocation *model.UserGameLocation)
	Find(userID uint) (*model.UserGameLocation, bool)
	Delete(userID uint)
	Update(userGameLocation *model.UserGameLocation)
	GetAllUserGameLocationsByRoomId(roomId uint) []*model.UserGameLocation
}
