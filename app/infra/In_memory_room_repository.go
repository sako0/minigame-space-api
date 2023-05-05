package infra

import (
	"sync"

	"github.com/sako0/minigame-space-api/app/domain/model"
)

type InMemoryRoomRepository struct {
	rooms map[uint]*model.Room
	mu    sync.RWMutex
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{rooms: make(map[uint]*model.Room)}
}

func (r *InMemoryRoomRepository) GetRoom(roomId uint) (*model.Room, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room, ok := r.rooms[roomId]
	return room, ok
}

func (r *InMemoryRoomRepository) AddRoom(roomId uint, room *model.Room) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rooms[roomId] = room
}

func (r *InMemoryRoomRepository) RemoveRoom(roomId uint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rooms, roomId)
}

func (r *InMemoryRoomRepository) GetAllRooms() []*model.Room {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rooms := make([]*model.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
