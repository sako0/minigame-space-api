package mysql

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
	"gorm.io/gorm"
)

type RoomRepository struct {
	// gormを使うのであれば、gorm.DBを埋め込む
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (repo *RoomRepository) StoreRoom(room *model.Room) {
	// repo.db.Exec("INSERT INTO rooms (id, name) VALUES (?, ?)", room.ID, room.Name)
	repo.db.Create(room)

}

func (repo *RoomRepository) LoadRoom(id string) (*model.Room, bool) {
	tx := repo.db.First(&model.Room{}, id)
	if tx.Error == nil {
		return &model.Room{}, true
	}
	return &model.Room{}, false
}

func (repo *RoomRepository) Delete(id string) {
	// データベースからルームを削除する処理
	repo.db.Delete(&model.Room{}, id)
}

func (repo *RoomRepository) ListRooms() []*model.Room {
	// データベースからルーム一覧を取得する処理
	repo.db.Find(&model.Room{})
	return []*model.Room{}
}
