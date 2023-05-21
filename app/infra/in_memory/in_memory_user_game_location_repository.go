package in_memory

import (
	"sync"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type InMemoryUserRoomLocationRepository struct {
	store map[uint]*model.UserGameLocation
	mu    sync.Mutex
}

func NewInMemoryUserGameLocationRepository() repository.InMemoryUserGameLocationRepository {
	return &InMemoryUserRoomLocationRepository{
		store: make(map[uint]*model.UserGameLocation),
	}
}

func (r *InMemoryUserRoomLocationRepository) Store(userRoomLocation *model.UserGameLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[userRoomLocation.UserID] = userRoomLocation
}

func (r *InMemoryUserRoomLocationRepository) Find(userID uint) (*model.UserGameLocation, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userRoomLocation, ok := r.store[userID]
	return userRoomLocation, ok
}

func (r *InMemoryUserRoomLocationRepository) Delete(userID uint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.store, userID)
}

func (r *InMemoryUserRoomLocationRepository) Update(userRoomLocation *model.UserGameLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[userRoomLocation.UserID] = userRoomLocation
}

func (r *InMemoryUserRoomLocationRepository) GetAllUserGameLocationsByRoomId(roomId uint) []*model.UserGameLocation {
	r.mu.Lock()
	defer r.mu.Unlock()

	userGameLocations := make([]*model.UserGameLocation, 0, len(r.store))
	for _, userGameLocation := range r.store {
		if userGameLocation.RoomID == roomId {
			userGameLocations = append(userGameLocations, userGameLocation)
		}
	}
	return userGameLocations
}
