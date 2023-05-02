package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type UserLocationRepository interface {
	GetUserLocation(userId string) (*model.UserLocation, error)
	AddUserLocation(userLocation *model.UserLocation) error
	UpdateUserLocation(userLocation *model.UserLocation) error
	RemoveUserLocation(userId string) error
	GetUserLocationsByRoom(roomId string) ([]model.UserLocation, error)
}
