package usecase

import (
	"log"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

// app/usecase/room_usecase.go
type RoomUsecase struct {
	roomRepo         repository.RoomRepository
	areaRepo         repository.AreaRepository
	userRepo         repository.UserRepository
	userLocationRepo repository.UserLocationRepository
	connectionStore  repository.ConnectionStoreRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository, areaRepo repository.AreaRepository, userRepo repository.UserRepository, userLocationRepo repository.UserLocationRepository, connectionStore repository.ConnectionStoreRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo, areaRepo: areaRepo, userRepo: userRepo, userLocationRepo: userLocationRepo, connectionStore: connectionStore}
}

func (uc *RoomUsecase) ConnectUserLocation(userLocation *model.UserLocation) (*model.UserLocation, error) {
	// 既に同じUserLocationが接続している場合は何もしない
	for _, otherUserLocation := range userLocation.Room.UserLocations {
		if otherUserLocation.RoomID == userLocation.Room.ID && otherUserLocation.UserID == userLocation.User.ID {
			return userLocation, nil
		}
	}
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return nil, err
	}
	return userLocation, nil
}

func (uc *RoomUsecase) DisconnectUserLocation(userLocation *model.UserLocation) {
	room, err := uc.roomRepo.GetRoom(userLocation.RoomID)
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
		uc.connectionStore.RemoveConnection(&userLocation.User) // コネクションを削除する
		userLocation.Conn.Close()
	}
}

func (uc *RoomUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) ([]uint, error) {
	userLocations, err := uc.userLocationRepo.GetUserLocationsByRoom(userLocation.RoomID)
	if err != nil {
		log.Printf("Error sending room joined event: %v", err)
		uc.DisconnectUserLocation(userLocation)
		return nil, err
	}
	var connectedUserIds []uint
	for _, ul := range userLocations {
		connectedUserIds = append(connectedUserIds, ul.UserID)
	}
	log.Println(connectedUserIds)
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"userId":           userLocation.UserID,
	}

	err = userLocation.Conn.WriteJSON(roomJoinedMsg)
	if err != nil {
		log.Printf("Error sending client-joined event to client: %v", err)
		uc.DisconnectUserLocation(userLocation)
	}

	return connectedUserIds, nil
}

func (uc *RoomUsecase) FindUserLocationInRoom(userId uint) (*model.UserLocation, error) {

	userLocation, err := uc.userLocationRepo.GetUserLocation(userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {

			return nil, nil
		}
		return nil, err
	}

	return userLocation, nil
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

func (uc *RoomUsecase) StoreConnection(user *model.User, conn *websocket.Conn) {
	uc.connectionStore.StoreConnection(user, conn)
}

func (uc *RoomUsecase) RemoveConnection(user *model.User) {
	uc.connectionStore.RemoveConnection(user)
}

func (uc *RoomUsecase) GetConnectionByUserID(userID uint) (*websocket.Conn, bool) {
	return uc.connectionStore.GetConnectionByUserID(userID)
}
