// app/infra/memory/connection_store.go
package infra

import (
	"sync"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type ConnectionStore struct {
	connections map[uint]*model.UserLocation
	mu          sync.RWMutex
}

func NewConnectionStore() repository.ConnectionStoreRepository {
	return &ConnectionStore{
		connections: make(map[uint]*model.UserLocation),
	}
}

func (s *ConnectionStore) StoreConnection(userLocation *model.UserLocation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[userLocation.User.ID] = userLocation
}

func (s *ConnectionStore) RemoveConnection(user *model.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, user.ID)
}
func (s *ConnectionStore) GetUserLocationByUserID(userID uint) (*model.UserLocation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userLocation, ok := s.connections[userID]
	return userLocation, ok
}

func (s *ConnectionStore) FindUserLocationInRoom(room *model.Room, userId uint) *model.UserLocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userLocation, ok := s.connections[userId]

	if ok && userLocation.Room.ID == room.ID {
		return userLocation
	}

	return nil
}
func (c *ConnectionStore) GetConnectedUserIdsInRoom(roomId uint) []uint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userIds := []uint{}
	for _, userLocation := range c.connections {
		if userLocation.Room.ID == roomId {
			userIds = append(userIds, userLocation.User.ID)
		}
	}

	return userIds
}
