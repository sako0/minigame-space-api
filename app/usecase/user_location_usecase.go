// app/usecase/room_usecase.go
package usecase

import (
	"encoding/json"
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
	_, exists, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		uc.DisconnectUserLocation(userLocation)
		return err
	}

	if !exists {
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
	_, exists, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		uc.DisconnectUserLocation(userLocation)
		return err
	}

	if !exists {
		log.Println("userLocation is not exist")
		err := uc.userLocationRepo.AddUserLocation(userLocation)
		if err != nil {
			uc.DisconnectUserLocation(userLocation)
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
		uc.DisconnectUserLocation(userLocation)
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
	userLocations, err := uc.GetSerializedConnectedUserLocations(userLocation.AreaID)
	if err != nil {
		return err
	}
	areaJoinedMsg := map[string]interface{}{
		"areaID":        userLocation.AreaID,
		"type":          "joined-area",
		"userLocations": userLocations,
		"fromUserID":    userLocation.UserID,
		"xAxis":         userLocation.XAxis,
		"yAxis":         userLocation.YAxis,
	}
	msg := model.NewMessage(areaJoinedMsg)
	return uc.SendMessageToSameArea(userLocation, msg)
}

func (uc *UserLocationUsecase) SendRoomJoinedEvent(userLocation *model.UserLocation) error {
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(userLocation.RoomID)
	connectedUserIds := []uint{}
	for _, otherUserLocation := range connectedUserLocations {
		connectedUserIds = append(connectedUserIds, otherUserLocation.UserID)
	}
	roomJoinedMsg := map[string]interface{}{
		"type":             "join-audio",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userLocation.UserID,
	}
	msg := model.NewMessage(roomJoinedMsg)
	return uc.SendMessageToSameRoom(userLocation, msg)
}

func (uc *UserLocationUsecase) MoveInArea(userLocation *model.UserLocation, xAxis int, yAxis int) error {
	userLocation.XAxis = xAxis
	userLocation.YAxis = yAxis
	log.Printf("XAxis: %d, YAxis: %d", xAxis, yAxis)
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	uc.inMemoryUserLocationRepo.Store(userLocation)
	userLocations, err := uc.GetSerializedConnectedUserLocations(userLocation.AreaID)
	if err != nil {
		return err
	}
	moveMsg := map[string]interface{}{
		"type":          "move",
		"areaID":        userLocation.AreaID,
		"fromUserID":    userLocation.UserID,
		"userLocations": userLocations,
		"xAxis":         userLocation.XAxis,
		"yAxis":         userLocation.YAxis,
	}

	msg := model.NewMessage(moveMsg)
	return uc.SendMessageToSameArea(userLocation, msg)
}

func (uc *UserLocationUsecase) SendMessageToSameArea(userLocation *model.UserLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	msgPayload["areaID"] = userLocation.AreaID
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByAreaId(userLocation.AreaID)
	for _, otherClient := range connectedUserLocations {
		err := otherClient.Conn.WriteJSON(msgPayload)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			uc.DisconnectUserLocation(otherClient)
			return err
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
			otherClient.Mutex.Lock()
			defer otherClient.Mutex.Unlock()
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
func (uc *UserLocationUsecase) SendMessageToSpecificUser(userLocation *model.UserLocation, msg *model.Message, targetUserID uint) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userLocation.UserID
	msgPayload["areaID"] = userLocation.AreaID
	msgPayload["roomID"] = userLocation.RoomID
	msgPayload["toUserID"] = targetUserID

	targetUserLocation, ok := uc.inMemoryUserLocationRepo.Find(targetUserID)
	if !ok {
		return fmt.Errorf("target user location not found for UserID: %d", targetUserID)
	}

	targetUserLocation.Mutex.Lock()
	defer targetUserLocation.Mutex.Unlock()
	err := targetUserLocation.Conn.WriteJSON(msgPayload)
	if err != nil {
		log.Printf("Error sending message to client: %v", err)
		uc.DisconnectUserLocation(targetUserLocation)
		return err
	}

	return nil
}

func (uc *UserLocationUsecase) LeaveInArea(userLocation *model.UserLocation) error {
	userLocation, ok, err := uc.userLocationRepo.GetUserLocation(userLocation.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user location not found")
	}
	userLocations, err := uc.GetSerializedConnectedUserLocations(userLocation.AreaID)
	if err != nil {
		return err
	}
	leaveMsg := map[string]interface{}{
		"type":          "leave-area",
		"areaID":        userLocation.AreaID,
		"roomID":        userLocation.RoomID,
		"fromUserID":    userLocation.UserID,
		"userLocations": userLocations,
	}
	msg := model.NewMessage(leaveMsg)
	uc.DisconnectUserLocation(userLocation)
	err = uc.userLocationRepo.RemoveUserLocation(userLocation.UserID)
	if err != nil {
		return fmt.Errorf("failed to remove user location: %w", err)
	}
	return uc.SendMessageToSameArea(userLocation, msg)
}

func (uc *UserLocationUsecase) LeaveInRoom(userLocation *model.UserLocation, roomID uint) error {
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserLocations {
		if otherClient.UserID != userLocation.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "leave-room",
				"areaID":     userLocation.AreaID,
				"roomID":     roomID,
				"fromUserID": userLocation.UserID,
				"toUserID":   otherClient.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := uc.SendMessageToSpecificUser(userLocation, msg, otherClient.UserID)
			if err != nil {
				return err
			}

		}
	}
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	err = uc.DisconnectUserLocation(userLocation)
	if err != nil {

		return err
	}
	return nil
}

func (uc *UserLocationUsecase) DisconnectInRoom(userLocation *model.UserLocation, roomID uint) error {
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserLocations {
		if otherClient.UserID != userLocation.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "disconnect-room",
				"areaID":     userLocation.AreaID,
				"roomID":     roomID,
				"fromUserID": userLocation.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := uc.SendMessageToSpecificUser(userLocation, msg, otherClient.UserID)
			if err != nil {
				return err
			}
		}
	}
	err := uc.userLocationRepo.UpdateUserLocation(userLocation)
	if err != nil {
		return err
	}
	err = uc.DisconnectUserLocation(userLocation)
	if err != nil {

		return err
	}
	return nil
}

func (uc *UserLocationUsecase) GetSerializedConnectedUserLocations(ariaID uint) ([]map[string]interface{}, error) {
	connectedUserLocations := uc.inMemoryUserLocationRepo.GetAllUserLocationsByAreaId(ariaID)
	userLocations := []map[string]interface{}{}
	for _, otherUserLocation := range connectedUserLocations {
		userLocation, exists, err := uc.userLocationRepo.GetUserLocation(otherUserLocation.UserID)

		if err != nil {
			uc.DisconnectUserLocation(userLocation)
			return nil, err
		}
		if !exists {
			log.Printf("user location does not exist for user ID: %d", otherUserLocation.UserID)
			err := uc.userLocationRepo.AddUserLocation(userLocation)
			if err != nil {
				uc.DisconnectUserLocation(userLocation)
				return nil, fmt.Errorf("failed to add user location: %w", err)
			}
		}
		userLocationJSON, err := userLocation.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user location from JSON: %w", err)

		}
		var userLocationMap map[string]interface{}
		err = json.Unmarshal(userLocationJSON, &userLocationMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user location from JSON: %w", err)
		}
		userLocations = append(userLocations, userLocationMap)
	}
	return userLocations, nil
}
