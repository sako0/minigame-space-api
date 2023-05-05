// app/usecase/room_usecase.go
package usecase

import (
	"fmt"
	"log"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type UserLocationUsecase struct {
	userLocationRepo         repository.UserLocationRepository
	inMemoryUserLocationRepo repository.InMemoryUserLocationRepository
}

func NewUserLocationUsecase(userLocationRepo repository.UserLocationRepository, inMemoryUserLocationRepo repository.InMemoryUserLocationRepository) *UserLocationUsecase {
	return &UserLocationUsecase{userLocationRepo: userLocationRepo, inMemoryUserLocationRepo: inMemoryUserLocationRepo}
}

func (uc *UserLocationUsecase) ConnectUserLocation(userLocation *model.UserLocation) (*model.UserLocation, error) {
	if userLocation.RoomID == 0 {
		return nil, fmt.Errorf("userLocation.RoomID is nil")
	}

	for _, otherUserLocation := range userLocation.Room.UserLocations {
		if otherUserLocation.RoomID == userLocation.RoomID && otherUserLocation.UserID == userLocation.UserID {
			return userLocation, nil
		}
	}
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return nil, err
	}

	uc.inMemoryUserLocationRepo.Store(userLocation)

	return userLocation, nil
}

func (uc *UserLocationUsecase) DisconnectUserLocation(userLocation *model.UserLocation) error {
	uc.inMemoryUserLocationRepo.Delete(userLocation.UserID)
	err := uc.userLocationRepo.RemoveUserLocation(userLocation.UserID)
	if err != nil {
		log.Printf("Error removing userLocation: %v", err)
		return err
	}
	return nil

}

func (uc *UserLocationUsecase) BroadcastMessage(userLocation *model.UserLocation, msgPayload map[string]interface{}) error {
	connectedUserIds := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)
	log.Printf("Initial connectedUserIds: %v", connectedUserIds)

	for _, otherClient := range connectedUserIds {
		if otherClient.UserID != userLocation.UserID {
			err := otherClient.Conn.WriteJSON(msgPayload)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				uc.DisconnectUserLocation(otherClient)
				return err
			}
		}
	}
	return nil
}

func (uc *UserLocationUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) ([]*model.UserLocation, error) {
	connectedUserIds := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)
	log.Printf("Initial connectedUserIds: %v", connectedUserIds)
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userLocation.UserID,
	}
	return connectedUserIds, uc.BroadcastMessage(userLocation, roomJoinedMsg)
}

func (uc *UserLocationUsecase) SendMessageToOtherClients(userLocation *model.UserLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	return uc.BroadcastMessage(userLocation, msgPayload)
}
