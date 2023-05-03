package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type AreaRepository struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) repository.AreaRepository {
	return &AreaRepository{db: db}
}

func (repo *AreaRepository) GetArea(areaId uint) (*model.Area, error) {
	area := &model.Area{}
	err := repo.db.First(area, areaId).Error
	if err != nil {
		return nil, err
	}
	return area, nil
}
