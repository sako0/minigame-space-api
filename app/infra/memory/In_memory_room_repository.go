package infra

import "github.com/sako0/minigame-space-api/app/domain/model"

type InMemoryRoomRepository struct {
	rooms map[string]*model.Room
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{rooms: make(map[string]*model.Room)}
}

func (repo *InMemoryRoomRepository) GetRoom(roomId string) (*model.Room, bool) {
	room, ok := repo.rooms[roomId]
	return room, ok
}

func (repo *InMemoryRoomRepository) AddRoom(roomId string, room *model.Room) {
	repo.rooms[roomId] = room
}

func (repo *InMemoryRoomRepository) RemoveRoom(roomId string) {
	delete(repo.rooms, roomId)
}
