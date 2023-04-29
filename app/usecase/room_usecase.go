package usecase

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
	"gorm.io/gorm"
)

type RoomUsecase struct {
	roomRepo         repository.RoomRepository
	areaRepo         repository.AreaRepository
	userRepo         repository.UserRepository
	userLocationRepo repository.UserLocationRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository, areaRepo repository.AreaRepository, userRepo repository.UserRepository, userLocationRepo repository.UserLocationRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo, areaRepo: areaRepo, userRepo: userRepo, userLocationRepo: userLocationRepo}
}

func (uc *RoomUsecase) ConnectUserLocation(room *model.Room, user *model.User, conn *websocket.Conn) (*model.UserLocation, error) {

	area, err := uc.areaRepo.GetArea(room.AreaID)
	if err != nil {
		return nil, err
	}
	userLocation, err := uc.userLocationRepo.GetUserLocation(user.ID)
	userLocation = &model.UserLocation{
		UserID: user.ID,
		AreaID: area.ID,
		RoomID: room.ID,
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {

			err = uc.userLocationRepo.AddUserLocation(userLocation)
			return userLocation, err
		}
		return nil, err
	}
	uc.userLocationRepo.UpdateUserLocation(userLocation)
	// 既に同じUserLocationが接続している場合は何もしない
	for _, otherUserLocation := range room.UserLocations {
		if otherUserLocation.RoomID == room.ID && otherUserLocation.UserID == user.ID {
			return nil, nil
		}
	}
	err = uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return nil, err
	}

	room.UserLocations = append(room.UserLocations, *userLocation)
	// ルームをデータベースに保存
	if err := uc.roomRepo.UpdateRoom(room); err != nil {
		return nil, err
	}
	return userLocation, nil
}

func (uc *RoomUsecase) DisconnectUserLocation(roomId uint, userLocation *model.UserLocation) {
	room, err := uc.roomRepo.GetRoom(roomId)
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
				uc.roomRepo.RemoveRoom(roomId)
			}
		}
	}
	if userLocation.Conn != nil {
		userLocation.Conn.Close()
	}
}
func (uc *RoomUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) ([]uint, error) {
	userLocations, err := uc.userLocationRepo.GetUserLocationsByRoom(userLocation.RoomID)
	if err != nil {
		return nil, err
	}

	var connectedUserIds []uint
	for _, ul := range userLocations {
		if ul.ID != userLocation.ID {
			connectedUserIds = append(connectedUserIds, ul.UserID)
			err := ul.Conn.WriteJSON(map[string]interface{}{
				"type":   "user-joined",
				"userId": userLocation.UserID,
			})
			if err != nil {
				return nil, errors.New("error sending user-joined event to other clients")
			}
		}
	}

	return connectedUserIds, nil
}

func (uc *RoomUsecase) FindUserLocationInRoom(roomId uint, userId uint) (*model.UserLocation, error) {
	room, err := uc.roomRepo.GetRoom(roomId)
	if err != nil {
		return nil, err
	}

	for _, userLocation := range room.UserLocations {
		if userLocation.UserID == userId {
			return &userLocation, nil
		}
	}

	return nil, fmt.Errorf("UserLocation not found for user %d in room %d", userId, roomId)
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
