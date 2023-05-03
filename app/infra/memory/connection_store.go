// app/infra/memory/connection_store.go
package infra

import (
	"fmt"
	"log"
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
	s.connections[userLocation.ID] = userLocation
}

func (s *ConnectionStore) RemoveConnection(userLocation *model.UserLocation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, userLocation.ID)
}
func (s *ConnectionStore) GetUserLocation(userLocatinId uint) (*model.UserLocation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userLocation, ok := s.connections[userLocatinId]
	return userLocation, ok
}

func (s *ConnectionStore) FindUserLocationInRoom(room *model.Room, userLocatinId uint) *model.UserLocation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userLocation, ok := s.connections[userLocatinId]

	if ok && userLocation.Room.ID == room.ID {
		return userLocation
	}

	return nil
}
func (c *ConnectionStore) GetConnectedUserIdsInRoom(roomId uint) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userIds := []string{}
	addedUserIds := map[string]bool{}
	for _, userLocation := range c.connections {
		firebaseUID := userLocation.User.FirebaseUID
		fmt.Printf("FirebaseUID: %#v\n", firebaseUID)

		if userLocation.Room.ID == roomId && !addedUserIds[firebaseUID] {
			userIds = append(userIds, firebaseUID)
			addedUserIds[firebaseUID] = true
		}
	}

	log.Println("userIds", userIds)

	return userIds
}
