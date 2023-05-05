package gorm

import (
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

func (r *UserLocationRepository) GetUserLocation(userId uint) (*model.UserLocation, error) {
	userLocation := &model.UserLocation{}
	result := r.db.First(userLocation, userId)
	return userLocation, result.Error
}

func (r *UserLocationRepository) AddUserLocation(userLocation *model.UserLocation) error {
	result := r.db.Create(userLocation)
	return result.Error
}

func (r *UserLocationRepository) RemoveUserLocation(userId uint) error {
	result := r.db.Delete(&model.UserLocation{}, userId)
	return result.Error
}

func (r *UserLocationRepository) UpdateUserLocation(userLocation *model.UserLocation) error {
	result := r.db.Save(userLocation)
	return result.Error
}

func (r *UserLocationRepository) GetAllUserLocationsByRoomId(roomId uint) ([]*model.UserLocation, error) {
	userLocations := []*model.UserLocation{}
	result := r.db.Where("room_id = ?", roomId).Find(&userLocations)
	return userLocations, result.Error
}
