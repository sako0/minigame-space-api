package gorm

import (
	"fmt"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type UserLocationRepository struct {
	db *gorm.DB
}

func NewUserLocationRepository(db *gorm.DB) repository.UserLocationRepository {
	return &UserLocationRepository{db: db}
}

func (r *UserLocationRepository) GetUserLocation(userId uint) (*model.UserLocation, bool, error) {
	userLocation := &model.UserLocation{}
	result := r.db.First(userLocation, userId)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetUserLocation: %v", result.Error)
	}

	return userLocation, true, nil
}

func (r *UserLocationRepository) AddUserLocation(userLocation *model.UserLocation) error {
	result := r.db.Create(userLocation)
	if result.Error != nil {
		return fmt.Errorf("AddUserLocation: %v", result.Error)
	}
	return nil
}

func (r *UserLocationRepository) RemoveUserLocation(userId uint) error {
	result := r.db.Delete(&model.UserLocation{}, userId)
	if result.Error != nil {
		return fmt.Errorf("RemoveUserLocation: %v", result.Error)
	}
	return nil
}

func (r *UserLocationRepository) UpdateUserLocation(userLocation *model.UserLocation) error {
	result := r.db.Save(userLocation)
	if result.Error != nil {
		return fmt.Errorf("UpdateUserLocation: %v", result.Error)
	}
	return nil
}

func (r *UserLocationRepository) GetAllUserLocationsByRoomId(roomId uint) ([]*model.UserLocation, bool, error) {
	userLocations := []*model.UserLocation{}
	result := r.db.Where("room_id = ?", roomId).Find(&userLocations)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetAllUserLocationsByRoomId: %v", result.Error)
	}

	return userLocations, true, nil
}
