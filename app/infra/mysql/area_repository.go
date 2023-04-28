package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type AreaRepository struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) (repository.AreaRepository, error) {
	return &AreaRepository{db: db}, nil
}
func (repo *AreaRepository) GetArea(areaId uint) (*model.Area, error) {
	var area model.Area
	err := repo.db.First(&area, "id = ?", areaId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &area, nil
}

func (repo *AreaRepository) AddArea(area *model.Area) error {
	err := repo.db.Create(area).Error
	if err != nil {
		return err
	}
	return nil
}

func (repo *AreaRepository) RemoveArea(areaId uint) error {
	err := repo.db.Delete(&model.Area{}, "id = ?", areaId).Error
	if err != nil {
		return err
	}
	return nil
}
