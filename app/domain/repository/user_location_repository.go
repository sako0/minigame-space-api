package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type UserLocationRepository interface {
	GetUserLocation(userId uint) (*model.UserLocation, bool, error)
	AddUserLocation(userLocation *model.UserLocation) error
	RemoveUserLocation(userId uint) error
	UpdateUserLocation(userLocation *model.UserLocation) error
	GetAllUserLocationsByAreaId(areaId uint) ([]*model.UserLocation, bool, error)
	GetAllUserLocationsByRoomId(roomId uint) ([]*model.UserLocation, bool, error)
}
