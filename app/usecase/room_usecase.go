// app/usecase/room_usecase.go
package usecase

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type RoomUsecase struct {
	roomRepo         repository.RoomRepository
	areaRepo         repository.AreaRepository
	userRepo         repository.UserRepository
	userLocationRepo repository.UserLocationRepository
	storeRepo        repository.ConnectionStoreRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository, areaRepo repository.AreaRepository, userRepo repository.UserRepository, userLocationRepo repository.UserLocationRepository, storeRepo repository.ConnectionStoreRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo, areaRepo: areaRepo, userRepo: userRepo, userLocationRepo: userLocationRepo, storeRepo: storeRepo}
}

func (uc *RoomUsecase) ConnectUserLocation(userLocation *model.UserLocation) (*model.UserLocation, error) {
	if userLocation.Room.ID == 0 {
		return nil, fmt.Errorf("userLocation.Room is nil")
	}

	for _, otherUserLocation := range userLocation.Room.UserLocations {
		fmt.Println("otherUserLocation :", otherUserLocation)
		fmt.Println("ConnectUserLocation  :", userLocation)
		if otherUserLocation.Room.ID == userLocation.Room.ID && otherUserLocation.User.ID == userLocation.User.ID {
			return userLocation, nil
		}
	}
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return nil, err
	}

	uc.storeRepo.StoreConnection(userLocation)

	return userLocation, nil
}

func (uc *RoomUsecase) DisconnectUserLocation(userLocation *model.UserLocation) {
	room, err := uc.roomRepo.GetRoom(userLocation.RoomID)
	if err != nil {
		log.Printf("Error disconnecting user location: %v", err)
		return
	}
	if err == nil {
		index := -1
		for i, ul := range room.UserLocations {
			if ul.ID == userLocation.ID {
				index = i
				break
			}
		}
		if index != -1 {
			room.UserLocations = append(room.UserLocations[:index], room.UserLocations[index+1:]...)
			if len(room.UserLocations) == 0 {
				uc.roomRepo.RemoveRoom(userLocation.RoomID)
			}
		}
	}
	if userLocation.Conn != nil {
		uc.storeRepo.RemoveConnection(userLocation) // Changed this line to use the connection store use case
		userLocation.Conn.Close()
	}
}

func (uc *RoomUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) ([]string, error) {
	userLocations, err := uc.userLocationRepo.GetUserLocationsByRoom(userLocation.Room.ID)
	if err != nil {
		log.Printf("Error sending room joined event: %v", err)
		uc.DisconnectUserLocation(userLocation)
		return nil, err
	}

	// 接続中のユーザーIDを取得する
	connectedUserIds := uc.storeRepo.GetConnectedUserIdsInRoom(userLocation.Room.ID)
	log.Printf("Initial connectedUserIds: %v", connectedUserIds)

	log.Println("userLocations :", userLocations)

	// for _, ul := range userLocations {
	// 	log.Printf("Checking ul: %#v", ul) // デバッグログを追加

	// 	// Skip the current user to prevent adding their own ID
	// 	if ul.User.ID != userLocation.User.ID {

	// 		// Check if the user has an active connection
	// 		if userLocation, ok := uc.storeRepo.GetUserLocation(ul.ID); ok {
	// 			connectedUserIds = append(connectedUserIds, userLocation.User.FirebaseUID)
	// 			log.Printf("Appended userId: %d, connectedUserIds: %v", userLocation.User.ID, connectedUserIds)
	// 		}
	// 	}
	// }

	log.Printf("Final connectedUserIds: %v", connectedUserIds)
	return connectedUserIds, nil
}

func (uc *RoomUsecase) HandleSignalMessage(userLocation *model.UserLocation, toUserConn *websocket.Conn, msg map[string]interface{}) {
	err := toUserConn.WriteJSON(msg)
	if err != nil {
		log.Printf("Error sending signal message to target user: %v", err)
		uc.DisconnectUserLocation(userLocation)
	}
}

func (uc *RoomUsecase) GetUserByFirebaseUID(firebaseUid string) (*model.User, error) {
	return uc.userRepo.GetUserByFirebaseUID(firebaseUid)
}

func (uc *RoomUsecase) GetArea(areaId uint) (*model.Area, error) {
	return uc.areaRepo.GetArea(areaId)
}

func (uc *RoomUsecase) GetRoom(roomId uint) (*model.Room, error) {
	return uc.roomRepo.GetRoom(roomId)
}

func (uc *RoomUsecase) GetUserLocationByUserID(userId uint) (*model.UserLocation, error) {
	return uc.userLocationRepo.GetUserLocationByUserID(userId)
}

func (uc *RoomUsecase) ExistUserLocation(userId uint) (bool, error) {
	return uc.userLocationRepo.ExistUserLocation(userId)
}

func (u *RoomUsecase) AddUserLocation(userLocation *model.UserLocation) error {
	return u.userLocationRepo.AddUserLocation(userLocation)
}
