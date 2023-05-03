// app/usecase/connection_store_usecase.go
package usecase

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type ConnectionStoreUsecase struct {
	connectionStoreRepo repository.ConnectionStoreRepository
	userLocationRepo    repository.UserLocationRepository
	userRepo            repository.UserRepository
	mutex               *sync.Mutex
}

func NewConnectionStoreUsecase(connectionStoreRepo repository.ConnectionStoreRepository, userLocationRepo repository.UserLocationRepository, userRepo repository.UserRepository) *ConnectionStoreUsecase {
	return &ConnectionStoreUsecase{connectionStoreRepo: connectionStoreRepo, userLocationRepo: userLocationRepo, userRepo: userRepo, mutex: &sync.Mutex{}}
}

func (cu *ConnectionStoreUsecase) StoreUserLocation(userLocation *model.UserLocation) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()

	cu.connectionStoreRepo.StoreConnection(userLocation)
}

func (cu *ConnectionStoreUsecase) RemoveConnection(userLocation *model.UserLocation) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()

	cu.connectionStoreRepo.RemoveConnection(userLocation)
}

func (cu *ConnectionStoreUsecase) GetConnectionByUserFirebaseUID(firebase_uid string) (*websocket.Conn, bool) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()
	user, err := cu.userRepo.GetUserByFirebaseUID(firebase_uid)
	if err != nil {
		return nil, false
	}
	userLocatin, err := cu.userLocationRepo.GetUserLocationByUserID(user.ID)
	if err != nil {
		return nil, false
	}

	userLocation, ok := cu.connectionStoreRepo.GetUserLocation(userLocatin.ID)

	if !ok {
		return nil, false
	}

	return userLocation.Conn, true
}
