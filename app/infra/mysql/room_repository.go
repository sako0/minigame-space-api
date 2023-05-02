package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) repository.RoomRepository {
	return &RoomRepository{db: db}
}

func (repo *RoomRepository) GetRoom(roomId string) (*model.Room, error) {
	var room model.Room
	err := repo.db.Where("id = ?", roomId).Preload("UserLocations").Find(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (repo *RoomRepository) AddRoom(room *model.Room) error {
	err := repo.db.Create(room).Error
	return err
}

func (repo *RoomRepository) RemoveRoom(roomId string) error {
	err := repo.db.Delete(model.Room{}, "id = ?", roomId).Error
	return err
}

func (repo *RoomRepository) UpdateRoom(room *model.Room) error {
	return repo.db.Model(&model.Room{}).Where("id = ?", room.ID).Updates(room).Error
}
