package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type ConnectionStoreRepository interface {
	StoreConnection(userLocation *model.UserLocation)
	RemoveConnection(userLocation *model.UserLocation)
	GetUserLocation(userLocatinId uint) (*model.UserLocation, bool)
	FindUserLocationInRoom(room *model.Room, userLocatinId uint) *model.UserLocation
	GetConnectedUserIdsInRoom(roomId uint) []string
}
