package infra

import "github.com/sako0/minigame-space-api/app/domain/model"

type InMemoryRoomRepository struct {
	rooms map[uint]*model.Room
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{rooms: make(map[uint]*model.Room)}
}

func (r *InMemoryRoomRepository) GetRoom(roomId uint) (*model.Room, bool) {
	room, ok := r.rooms[roomId]
	return room, ok
}

func (r *InMemoryRoomRepository) AddRoom(roomId uint, room *model.Room) {
	r.rooms[roomId] = room
}

func (r *InMemoryRoomRepository) RemoveRoom(roomId uint) {
	delete(r.rooms, roomId)
}

func (r *InMemoryRoomRepository) GetAllRooms() []*model.Room {
	rooms := make([]*model.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
