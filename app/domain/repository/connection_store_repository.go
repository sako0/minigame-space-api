package repository

import (
	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type ConnectionStoreRepository interface {
	StoreConnection(user *model.User, conn *websocket.Conn)
	RemoveConnection(user *model.User)
	GetConnectionByUserID(userID uint) (*websocket.Conn, bool)
	FindUserLocationInRoom(room *model.Room, userId uint) *model.UserLocation
}
