package infra

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) (repository.UserRepository, error) {
	return &UserRepository{db: db}, nil
}

func (repo *UserRepository) GetUser(firebaseUid string) (*model.User, error) {
	user := model.User{}
	if err := repo.db.First(&user, "firebase_uid = ?", firebaseUid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) AddUser(user *model.User) error {
	err := repo.db.Create(user).Error
	if err != nil {

		return err
	}
	return nil
}

func (repo *UserRepository) UpdateUser(user *model.User) error {
	err := repo.db.Save(user).Error
	if err != nil {
		return err
	}
	return nil
}
