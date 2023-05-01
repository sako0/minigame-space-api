package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomRepository interface {
	GetRoom(roomId string) (*model.Room, error)
	AddRoom(room *model.Room) error
	RemoveRoom(roomId string) error
	UpdateRoom(room *model.Room) error
}
