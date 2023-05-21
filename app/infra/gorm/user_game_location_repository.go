package gorm

import (
	"fmt"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type UserGameLocationRepository struct {
	db *gorm.DB
}

func NewUserGameLocationRepository(db *gorm.DB) repository.UserGameLocationRepository {
	return &UserGameLocationRepository{db: db}
}

func (r *UserGameLocationRepository) GetUserGameLocation(userId uint) (*model.UserGameLocation, bool, error) {
	userGameLocation := &model.UserGameLocation{}
	result := r.db.First(userGameLocation, fmt.Sprintf("user_id = %d", userId))

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetUserGameLocation: %v", result.Error)
	}

	return userGameLocation, true, nil
}

func (r *UserGameLocationRepository) AddUserGameLocation(userGameLocation *model.UserGameLocation) error {
	result := r.db.Create(userGameLocation)
	if result.Error != nil {
		return fmt.Errorf("AddUserGameLocation: %v", result.Error)
	}
	return nil
}

func (r *UserGameLocationRepository) RemoveUserGameLocation(userId uint) error {
	result := r.db.Unscoped().Delete(&model.UserGameLocation{}, fmt.Sprintf("user_id = %d", userId))
	if result.Error != nil {
		return fmt.Errorf("RemoveUserGameLocation: %v", result.Error)
	}
	return nil
}

func (r *UserGameLocationRepository) UpdateUserGameLocation(userGameLocation *model.UserGameLocation) error {
	result := r.db.Where("user_id = ?", userGameLocation.UserID).Updates(userGameLocation)
	if result.Error != nil {
		return fmt.Errorf("UpdateUserGameLocation: %v", result.Error)
	}
	return nil
}

func (r *UserGameLocationRepository) GetAllUserGameLocationsByRoomId(roomId uint) ([]*model.UserGameLocation, bool, error) {
	userGameLocations := []*model.UserGameLocation{}
	result := r.db.Where("room_id = ?", roomId).Find(&userGameLocations)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetAllUserGameLocationsByRoomId: %v", result.Error)
	}

	return userGameLocations, true, nil
}
