package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomRepository interface {
	StoreRoom(room *model.Room)
	LoadRoom(id uint) (*model.Room, bool)
	Delete(id uint)
	ListRooms() []*model.Room
}
