package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type RoomTypeRepository interface {
	GetRoomType(id uint) (*model.RoomType, error)
	GetRoomTypes() ([]*model.RoomType, error)
}
