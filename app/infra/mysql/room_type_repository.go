package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type RoomTypeRepository struct {
	db *gorm.DB
}

func NewRoomTypeRepository(db *gorm.DB) (repository.RoomTypeRepository, error) {
	return &RoomTypeRepository{db: db}, nil
}

func (repo *RoomTypeRepository) GetRoomType(id uint) (*model.RoomType, error) {
	var roomType model.RoomType
	if err := repo.db.First(&roomType, id).Error; err != nil {
		return nil, err
	}
	return &roomType, nil
}

func (repo *RoomTypeRepository) GetRoomTypes() ([]*model.RoomType, error) {
	var roomTypes []*model.RoomType
	if err := repo.db.Find(&roomTypes).Error; err != nil {
		return nil, err
	}
	return roomTypes, nil
}
