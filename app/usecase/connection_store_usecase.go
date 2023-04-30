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
	mutex               *sync.Mutex
}

func NewConnectionStoreUsecase(connectionStoreRepo repository.ConnectionStoreRepository) *ConnectionStoreUsecase {
	return &ConnectionStoreUsecase{connectionStoreRepo: connectionStoreRepo, mutex: &sync.Mutex{}}
}

func (cu *ConnectionStoreUsecase) StoreUserLocation(userLocation *model.UserLocation) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()

	cu.connectionStoreRepo.StoreConnection(userLocation)
}

func (cu *ConnectionStoreUsecase) RemoveConnection(user *model.User) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()

	cu.connectionStoreRepo.RemoveConnection(user)
}

func (cu *ConnectionStoreUsecase) GetConnectionByUserID(userID uint) (*websocket.Conn, bool) {
	cu.mutex.Lock()
	defer cu.mutex.Unlock()

	userLocation, ok := cu.connectionStoreRepo.GetUserLocationByUserID(userID)

	if ok {
		return userLocation.Conn, ok
	}

	return nil, false
}
