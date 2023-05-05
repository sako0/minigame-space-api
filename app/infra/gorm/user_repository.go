package gorm

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

func (r *UserRepository) GetUser(userId uint) (*model.User, error) {
	user := &model.User{}
	result := r.db.First(user, userId)
	return user, result.Error
}

func (r *UserRepository) AddUser(user *model.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *UserRepository) RemoveUser(userId uint) error {
	result := r.db.Delete(&model.User{}, userId)
	return result.Error
}
func (r *UserRepository) GetUserByFirebaseUID(firebaseUID string) (*model.User, error) {
	user := &model.User{}
	result := r.db.Where("firebase_uid = ?", firebaseUID).First(user)
	return user, result.Error
}
