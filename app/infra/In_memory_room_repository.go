package infra

import "github.com/sako0/minigame-space-api/app/domain/model"

type InMemoryRoomRepository struct {
	rooms map[uint]*model.Room
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{rooms: make(map[uint]*model.Room)}
}

func (repo *InMemoryRoomRepository) GetRoom(roomId uint) (*model.Room, bool) {
	room, ok := repo.rooms[roomId]
	return room, ok
}

func (repo *InMemoryRoomRepository) AddRoom(roomId uint, room *model.Room) {
	repo.rooms[roomId] = room
}

func (repo *InMemoryRoomRepository) RemoveRoom(roomId uint) {
	delete(repo.rooms, roomId)
}
