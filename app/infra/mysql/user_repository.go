package infra

import (
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

func (repo *UserRepository) GetUser(userId uint) (*model.User, error) {
	user := &model.User{}
	err := repo.db.First(user, userId).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetUserByFirebaseUID(firebaseUID string) (*model.User, error) {
	user := &model.User{}
	err := repo.db.Where("firebase_uid = ?", firebaseUID).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
