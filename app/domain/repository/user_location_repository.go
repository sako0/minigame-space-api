package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type UserLocationRepository interface {
	GetUserLocation(userId uint) (*model.UserLocation, error)
	AddUserLocation(userLocation *model.UserLocation) error
	RemoveUserLocation(userId uint) error
	UpdateUserLocation(userLocation *model.UserLocation) error
	GetAllUserLocationsByRoomId(roomId uint) ([]*model.UserLocation, error)
}
