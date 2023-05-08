// app/usecase/room_usecase.go
package usecase

import (
	"errors"
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

func (uc *UserLocationUsecase) ConnectUserLocationForArea(userLocation *model.UserLocation) error {
	if userLocation.AreaID == 0 {
		return fmt.Errorf("userLocation.AreaID is nil")
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

	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByAreaId(userLocation.AreaID)

	// connectedUserLocationsの中に既に同じAreaIDでかつ同じUserIDを持つものがある場合は何もせずに終了
	for _, otherUserLocation := range connectedUserLocations {
		if otherUserLocation.AreaID == userLocation.AreaID && otherUserLocation.UserID == userLocation.UserID {
			return nil
		}
	}
	err = uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	uc.inMemoryUserLocationRepo.Store(userLocation)

	return nil
}
func (uc *UserLocationUsecase) ConnectUserLocationForRoom(userLocation *model.UserLocation) error {
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
	uc.inMemoryUserLocationRepo.Store(userLocation)

	return nil
}

func (uc *UserLocationUsecase) DisconnectUserLocation(userLocation *model.UserLocation) error {
	uc.inMemoryUserLocationRepo.Delete(userLocation.UserID)

	return nil
}

func (uc *UserLocationUsecase) SendAreaJoinedEvent(userLocation *model.UserLocation) error {
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByAreaId(userLocation.AreaID)

	connectedUserIds := []uint{}
	for _, otherUserLocation := range connectedUserLocations {
		if otherUserLocation.UserID != userLocation.UserID {
			connectedUserIds = append(connectedUserIds, otherUserLocation.UserID)
		}
	}
	log.Printf("Initial connectedUserIds: %v", connectedUserIds)
	areaJoinedMsg := map[string]interface{}{
		"areaId":           userLocation.AreaID,
		"type":             "joined-area",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userLocation.UserID,
	}
	msg := model.NewMessage(areaJoinedMsg)
	return uc.SendMessageToSameArea(userLocation, msg)
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
	msg := model.NewMessage(roomJoinedMsg)
	return uc.SendMessageToSameRoom(userLocation, msg)
}

func (uc *UserLocationUsecase) UpdateUserLocationAndBroadcastInArea(userLocation *model.UserLocation) error {
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	moveMsg := map[string]interface{}{
		"type":       "move",
		"AreaID":     userLocation.AreaID,
		"fromUserID": userLocation.UserID,
		"xAxis":      userLocation.XAxis,
		"yAxis":      userLocation.YAxis,
	}

	msg := model.NewMessage(moveMsg)
	return uc.SendMessageToSameArea(userLocation, msg)
}

func (uc *UserLocationUsecase) SendMessageToSameArea(userLocation *model.UserLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	msgPayload["areaId"] = userLocation.AreaID
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByAreaId(userLocation.AreaID)
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
func (uc *UserLocationUsecase) SendMessageToSameRoom(userLocation *model.UserLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	msgPayload["roomID"] = userLocation.RoomID
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)
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

func (uc *UserLocationUsecase) LeaveInArea(userLocation *model.UserLocation) error {
	leaveMsg := map[string]interface{}{
		"type":       "leave-area",
		"areaID":     userLocation.AreaID,
		"roomID":     userLocation.RoomID,
		"fromUserID": userLocation.UserID,
	}
	msg := model.NewMessage(leaveMsg)
	return uc.SendMessageToSameArea(userLocation, msg)
}

func (uc *UserLocationUsecase) LeaveInRoom(userLocation *model.UserLocation) error {
	userLocation, ok, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user location not found")
	}

	leaveMsg := map[string]interface{}{
		"type":       "leave-room",
		"areaId":     userLocation.AreaID,
		"roomId":     userLocation.RoomID,
		"fromUserID": userLocation.UserID,
	}
	msg := model.NewMessage(leaveMsg)
	return uc.SendMessageToSameRoom(userLocation, msg)
}

func (uc *UserLocationUsecase) DisconnectInAll(userLocation *model.UserLocation) error {
	userLocation, ok, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user location not found")
	}
	disconnectMsg := map[string]interface{}{
		"type":       "disconnect",
		"areaId":     userLocation.AreaID,
		"roomId":     userLocation.RoomID,
		"fromUserID": userLocation.UserID,
	}
	msg := model.NewMessage(disconnectMsg)
	uc.inMemoryUserLocationRepo.Delete(userLocation.UserID)
	// エリア内にルームがあるので、エリア内のユーザーに送信すれば良い
	return uc.SendMessageToSameArea(userLocation, msg)
}
