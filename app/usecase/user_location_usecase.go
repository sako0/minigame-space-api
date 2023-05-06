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

func (uc *UserLocationUsecase) ConnectUserLocation(userLocation *model.UserLocation) error {
	if userLocation.RoomID == 0 {
		return fmt.Errorf("userLocation.RoomID is nil")
	}

	// UserLocationが存在しない場合は新規作成
	_, isExist, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		return err
	}

	if !isExist {
		log.Println("userLocation is not exist")
		err := uc.userLocationRepo.AddUserLocation(userLocation)
		if err != nil {
			return err
		}
	}

	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)

	// connectedUserLocationsの中に既に同じRoomIDでかつ同じUserIDを持つものがある場合は何もせずに終了
	for _, otherUserLocation := range connectedUserLocations {
		if otherUserLocation.RoomID == userLocation.RoomID && otherUserLocation.UserID == userLocation.UserID {
			return nil
		}
	}
	err = uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	log.Println("userLocation: ", userLocation)
	uc.inMemoryUserLocationRepo.Store(userLocation)

	return nil
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
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)

	log.Printf("Initial connectedUserLocations: %v", connectedUserLocations)

	for _, otherClient := range connectedUserLocations {
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

func (uc *UserLocationUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) error {

	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)

	connectedUserIds := []uint{}
	for _, otherUserLocation := range connectedUserLocations {
		if otherUserLocation.UserID == userLocation.UserID {
			connectedUserIds = append(connectedUserIds, otherUserLocation.UserID)
		}
	}
	log.Printf("Initial connectedUserIds: %v", connectedUserIds)
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userLocation.UserID,
	}
	return uc.BroadcastMessage(userLocation, roomJoinedMsg)
}

func (uc *UserLocationUsecase) SendMessageToOtherClients(userLocation *model.UserLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	return uc.BroadcastMessage(userLocation, msgPayload)
}
