package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomRepository interface {
	GetRoom(roomId string) (*model.Room, bool)
	AddRoom(roomId string, room *model.Room)
	RemoveRoom(roomId string)
}
