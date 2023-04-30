package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type ConnectionStoreRepository interface {
	StoreConnection(userLocation *model.UserLocation)
	RemoveConnection(user *model.User)
	GetUserLocationByUserID(userID uint) (*model.UserLocation, bool)
	FindUserLocationInRoom(room *model.Room, userId uint) *model.UserLocation
	GetConnectedUserIdsInRoom(roomId uint) []uint
}
