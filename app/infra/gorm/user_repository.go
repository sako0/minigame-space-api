package gorm

import (
	"fmt"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(userId uint) (*model.User, bool, error) {
	user := &model.User{}
	result := r.db.First(user, userId)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetUser: %v", result.Error)
	}

	return user, true, nil
}

func (r *UserRepository) AddUser(user *model.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("AddUser: %v", result.Error)
	}
	return nil
}

func (r *UserRepository) RemoveUser(userId uint) error {
	result := r.db.Delete(&model.User{}, userId)
	if result.Error != nil {
		return fmt.Errorf("RemoveUser: %v", result.Error)
	}
	return nil
}

func (r *UserRepository) GetUserByFirebaseUID(firebaseUID string) (*model.User, bool, error) {
	user := &model.User{}
	result := r.db.Where("firebase_uid = ?", firebaseUID).First(user)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if result.Error != nil {
		return nil, false, fmt.Errorf("GetUserByFirebaseUID: %v", result.Error)
	}

	return user, true, nil
}
