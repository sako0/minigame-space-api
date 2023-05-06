package in_memory

import (
	"sync"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type InMemoryUserLocationRepository struct {
	store map[uint]*model.UserLocation // Key: userID, Value: UserLocation
	mu    sync.Mutex
}

func NewInMemoryUserLocationRepository() repository.InMemoryUserLocationRepository {
	return &InMemoryUserLocationRepository{
		store: make(map[uint]*model.UserLocation),
	}
}

func (r *InMemoryUserLocationRepository) Store(userLocation *model.UserLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[userLocation.UserID] = userLocation
}

func (r *InMemoryUserLocationRepository) Find(userID uint) (*model.UserLocation, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userLocation, ok := r.store[userID]
	return userLocation, ok
}

func (r *InMemoryUserLocationRepository) Delete(userID uint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.store, userID)
}

func (r *InMemoryUserLocationRepository) Update(userLocation *model.UserLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[userLocation.UserID] = userLocation
}

func (r *InMemoryUserLocationRepository) GetAllUserLocationsByRoomId(roomId uint) []*model.UserLocation {
	r.mu.Lock()
	defer r.mu.Unlock()

	userLocations := make([]*model.UserLocation, 0, len(r.store))
	for _, userLocation := range r.store {
		if userLocation.RoomID == roomId {
			userLocations = append(userLocations, userLocation)
		}
	}
	return userLocations
}
