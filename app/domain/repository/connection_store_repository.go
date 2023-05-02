package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type ConnectionStoreRepository interface {
	StoreConnection(userLocation *model.UserLocation)
	RemoveConnection(user *model.User)
	GetUserLocationByUserID(userID string) (*model.UserLocation, bool)
	FindUserLocationInRoom(room *model.Room, userId string) *model.UserLocation
	GetConnectedUserIdsInRoom(roomId string) []string
}
