package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomRepository interface {
	GetRoom(roomId uint) (*model.Room, bool)
	AddRoom(roomId uint, room *model.Room)
	RemoveRoom(roomId uint)
}
