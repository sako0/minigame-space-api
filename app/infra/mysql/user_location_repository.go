package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type UserLocationRepository struct {
	db *gorm.DB
}

func NewUserLocationRepository(db *gorm.DB) (repository.UserLocationRepository, error) {
	return &UserLocationRepository{db}, nil
}

func (r *UserLocationRepository) GetUserLocationById(id uint) (*model.UserLocation, error) {
	var userLocation model.UserLocation
	if err := r.db.First(&userLocation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &userLocation, nil
}
func (r *UserLocationRepository) SaveUserLocation(userLocation *model.UserLocation) error {
	var existingRecord model.UserLocation
	if err := r.db.Where("user_id = ? AND area_id = ? AND room_id = ?", userLocation.UserID, userLocation.AreaID, userLocation.RoomID).First(&existingRecord).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	if existingRecord.ID != 0 {
		// Update the existing record
		userLocation.ID = existingRecord.ID
	}

	return r.db.Save(userLocation).Error
}

func (r *UserLocationRepository) DeleteUserLocationById(id uint) error {
	return r.db.Delete(&model.UserLocation{}, id).Error
}

func (r *UserLocationRepository) AddUserLocation(userLocation *model.UserLocation) error {
	return r.db.Create(userLocation).Error
}
