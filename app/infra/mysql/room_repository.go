package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) (repository.RoomRepository, error) {
	return &RoomRepository{db: db}, nil
}
func (repo *RoomRepository) GetRoom(roomId uint) (*model.Room, error) {
	var room model.Room
	err := repo.db.First(&room, "id = ?", roomId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (repo *RoomRepository) AddRoom(room *model.Room) error {
	err := repo.db.Create(room).Error
	if err != nil {
		return err
	}
	return nil

}
func (repo *RoomRepository) RemoveRoom(roomId uint) error {
	err := repo.db.Delete(&model.Room{}, "id = ?", roomId).Error
	if err != nil {
		return err
	}
	return nil
}
