package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type InMemoryUserLocationRepository interface {
	Store(userLocation *model.UserLocation)
	Find(userID uint) (*model.UserLocation, bool)
	Delete(userID uint)
	Update(userLocation *model.UserLocation)
	GetAllUserLocationsByAreaId(areaId uint) []*model.UserLocation
	GetAllUserLocationsByRoomId(roomId uint) []*model.UserLocation
}
