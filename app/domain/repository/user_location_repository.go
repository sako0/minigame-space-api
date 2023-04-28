package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type UserLocationRepository interface {
	GetUserLocationById(id uint) (*model.UserLocation, error)
	SaveUserLocation(userLocation *model.UserLocation) error
	DeleteUserLocationById(id uint) error
	AddUserLocation(userLocation *model.UserLocation) error
}
