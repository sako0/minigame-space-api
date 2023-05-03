package infra

import (
	"log"

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

func (repo *UserLocationRepository) GetUserLocationByUserID(userId uint) (*model.UserLocation, error) {
	var userLocation model.UserLocation
	err := repo.db.Preload("User").Preload("Area").Preload("Room").Where("user_id = ?", userId).Find(&userLocation).Error
	if err != nil {
		return nil, err
	}
	log.Println("GetUserLocationByUserID userLocation :", userLocation)
	return &userLocation, nil
}

func (repo *UserLocationRepository) AddUserLocation(userLocation *model.UserLocation) error {
	log.Println("AddUserLocation :", userLocation)
	err := repo.db.Create(userLocation).Error
	return err
}

func (repo *UserLocationRepository) UpdateUserLocation(userLocation *model.UserLocation) error {
	log.Println("UpdateUserLocation :", userLocation)
	err := repo.db.Model(userLocation).Where("user_id = ?", userLocation.UserID).Updates(userLocation).Error
	return err
}

func (repo *UserLocationRepository) RemoveUserLocation(userId uint) error {
	err := repo.db.Delete(model.UserLocation{}, "user_id = ?", userId).Error
	return err
}

func (repo *UserLocationRepository) GetUserLocationsByRoom(roomId uint) ([]model.UserLocation, error) {
	var userLocations []model.UserLocation
	err := repo.db.Preload("User").Preload("Area").Preload("Room").Where("room_id = ?", roomId).Find(&userLocations).Error
	if err != nil {
		return nil, err
	}
	return userLocations, nil
}

func (repo *UserLocationRepository) ExistUserLocation(userId uint) (bool, error) {
	var count int64
	err := repo.db.Model(model.UserLocation{}).Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
