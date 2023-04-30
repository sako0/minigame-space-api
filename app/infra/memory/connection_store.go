// app/infra/memory/connection_store.go
package infra

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type ConnectionStore struct {
	connections map[uint]*websocket.Conn
	mu          sync.RWMutex
}

func NewConnectionStore() repository.ConnectionStoreRepository {
	return &ConnectionStore{
		connections: make(map[uint]*websocket.Conn),
	}
}

func (s *ConnectionStore) StoreConnection(user *model.User, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[user.ID] = conn
}

func (s *ConnectionStore) RemoveConnection(user *model.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, user.ID)
}

func (s *ConnectionStore) GetConnectionByUserID(userID uint) (*websocket.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, ok := s.connections[userID]
	return conn, ok
}

func (s *ConnectionStore) FindUserLocationInRoom(room *model.Room, userId uint) *model.UserLocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userLocation *model.UserLocation

	if userLocation != nil && userLocation.UserID == userId && s.connections[userLocation.UserID] != nil {
		userLocation = &model.UserLocation{
			UserID: userId,
		}
	}

	return userLocation
}
