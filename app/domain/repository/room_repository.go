package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomRepository interface {
	GetRoom(roomId uint) (*model.Room, error)
	AddRoom(room *model.Room) error
	RemoveRoom(roomId uint) error
	UpdateRoom(room *model.Room) error
}
