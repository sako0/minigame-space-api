// app/usecase/room_usecase.go
package usecase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type UserGameLocationUsecase struct {
	userGameLocationRepo         repository.UserGameLocationRepository
	inMemoryUserGameLocationRepo repository.InMemoryUserGameLocationRepository
}

func NewUserGameLocationUsecase(userGameLocationRepo repository.UserGameLocationRepository, inMemoryUserGameLocationRepo repository.InMemoryUserGameLocationRepository) *UserGameLocationUsecase {
	return &UserGameLocationUsecase{userGameLocationRepo: userGameLocationRepo, inMemoryUserGameLocationRepo: inMemoryUserGameLocationRepo}
}

func (ugc *UserGameLocationUsecase) ConnectUserGameLocation(userGameLocation *model.UserGameLocation) error {
	if userGameLocation.RoomID == 0 {
		return fmt.Errorf("userGameLocation.RoomID is nil")
	}

	// UserGameLocationが存在しない場合は新規作成
	_, exists, err := ugc.userGameLocationRepo.GetUserGameLocation(userGameLocation.UserID)
	if err != nil {
		ugc.DisconnectUserGameLocation(userGameLocation)
		log.Println("failed to get userGameLocation")
		return err
	}

	if !exists {
		log.Println("userGameLocation is not exist")
		err := ugc.userGameLocationRepo.AddUserGameLocation(userGameLocation)
		if err != nil {
			ugc.DisconnectUserGameLocation(userGameLocation)
			return err
		}
	}

	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(userGameLocation.RoomID)

	for _, otherUserGameLocation := range connectedUserGameLocations {
		if otherUserGameLocation.RoomID == userGameLocation.RoomID && otherUserGameLocation.UserID == userGameLocation.UserID {
			return nil
		}
	}
	err = ugc.userGameLocationRepo.UpdateUserGameLocation(userGameLocation)
	if err != nil {
		ugc.DisconnectUserGameLocation(userGameLocation)
		return err
	}
	ugc.inMemoryUserGameLocationRepo.Store(userGameLocation)

	return nil
}

func (ugc *UserGameLocationUsecase) DisconnectUserGameLocation(userGameLocation *model.UserGameLocation) error {
	log.Println("disconnect userGameLocation:", userGameLocation.UserID)
	ugc.inMemoryUserGameLocationRepo.Delete(userGameLocation.UserID)

	return nil
}

func (ugc *UserGameLocationUsecase) SendGameJoinedEvent(userGameLocation *model.UserGameLocation) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(userGameLocation.RoomID)
	connectedUserIds := []uint{}
	for _, otherUserGameLocation := range connectedUserGameLocations {
		connectedUserIds = append(connectedUserIds, otherUserGameLocation.UserID)
	}
	userLocations, err := ugc.GetSerializedConnectedUserGameLocations(userGameLocation.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get serialized connected user locations: %w", err)
	}
	roomJoinedMsg := map[string]interface{}{
		"type":              "join-game",
		"connectedUserIds":  connectedUserIds,
		"fromUserID":        userGameLocation.UserID,
		"xAxis":             userGameLocation.XAxis,
		"yAxis":             userGameLocation.YAxis,
		"roomID":            userGameLocation.RoomID,
		"userGameLocations": userLocations,
	}
	msg := model.NewMessage(roomJoinedMsg)
	return ugc.SendMessageToSameRoom(userGameLocation, msg)
}

func (ugc *UserGameLocationUsecase) SendAudioJoinedEvent(userGameLocation *model.UserGameLocation) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(userGameLocation.RoomID)
	connectedUserIds := []uint{}
	for _, otherUserGameLocation := range connectedUserGameLocations {
		connectedUserIds = append(connectedUserIds, otherUserGameLocation.UserID)
	}
	roomJoinedMsg := map[string]interface{}{
		"type":             "join-audio",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userGameLocation.UserID,
		"roomID":           userGameLocation.RoomID,
	}
	msg := model.NewMessage(roomJoinedMsg)
	return ugc.SendMessageToSameRoomWithoutMe(userGameLocation, msg)
}

func (ugc *UserGameLocationUsecase) SendMessageToSameRoomWithoutMe(userGameLocation *model.UserGameLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userGameLocation.UserID
	msgPayload["roomID"] = userGameLocation.RoomID
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(userGameLocation.RoomID)
	for _, otherClient := range connectedUserGameLocations {
		if otherClient.UserID != userGameLocation.UserID {
			otherClient.Mutex.Lock()
			defer otherClient.Mutex.Unlock()
			err := otherClient.Conn.WriteJSON(msgPayload)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				ugc.DisconnectUserGameLocation(otherClient)
				return err
			}
		}
	}
	return nil
}

func (ugc *UserGameLocationUsecase) SendMessageToSameRoom(userGameLocation *model.UserGameLocation, msg *model.Message) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userGameLocation.UserID
	msgPayload["roomID"] = userGameLocation.RoomID
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(userGameLocation.RoomID)
	for _, otherClient := range connectedUserGameLocations {
		otherClient.Mutex.Lock()
		defer otherClient.Mutex.Unlock()
		err := otherClient.Conn.WriteJSON(msgPayload)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			ugc.DisconnectUserGameLocation(otherClient)
			return err
		}

	}
	return nil
}

func (ugc *UserGameLocationUsecase) SendMessageToSpecificUser(userGameLocation *model.UserGameLocation, msg *model.Message, targetUserID uint) error {
	msgPayload := msg.Payload
	msgPayload["fromUserID"] = userGameLocation.UserID
	msgPayload["roomID"] = userGameLocation.RoomID
	msgPayload["toUserID"] = targetUserID

	targetUserGameLocation, ok := ugc.inMemoryUserGameLocationRepo.Find(targetUserID)
	if !ok {
		return fmt.Errorf("target user location not found for UserID: %d", targetUserID)
	}

	targetUserGameLocation.Mutex.Lock()
	defer targetUserGameLocation.Mutex.Unlock()
	err := targetUserGameLocation.Conn.WriteJSON(msgPayload)
	if err != nil {
		log.Printf("Error sending message to client: %v", err)
		ugc.DisconnectUserGameLocation(targetUserGameLocation)
		return err
	}

	return nil
}

func (ugc *UserGameLocationUsecase) MoveInGame(userGameLocation *model.UserGameLocation, xAxis int, yAxis int) error {
	userGameLocation.XAxis = xAxis
	userGameLocation.YAxis = yAxis
	err := ugc.userGameLocationRepo.UpdateUserGameLocation(userGameLocation)
	if err != nil {
		return err
	}
	ugc.inMemoryUserGameLocationRepo.Store(userGameLocation)
	userGameLocations, err := ugc.GetSerializedConnectedUserGameLocations(userGameLocation.RoomID)
	if err != nil {
		return err
	}
	moveMsg := map[string]interface{}{
		"type":              "move",
		"fromUserID":        userGameLocation.UserID,
		"roomID":            userGameLocation.RoomID,
		"userGameLocations": userGameLocations,
	}
	msg := model.NewMessage(moveMsg)
	err = ugc.SendMessageToSameRoom(userGameLocation, msg)
	if err != nil {
		return err
	}
	return nil
}

func (ugc *UserGameLocationUsecase) LeaveInGame(userGameLocation *model.UserGameLocation, roomID uint) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserGameLocations {
		if otherClient.UserID != userGameLocation.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "leave-game",
				"roomID":     roomID,
				"fromUserID": userGameLocation.UserID,
				"toUserID":   otherClient.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := ugc.SendMessageToSpecificUser(userGameLocation, msg, otherClient.UserID)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				return err
			}

		}
	}
	err := ugc.userGameLocationRepo.UpdateUserGameLocation(userGameLocation)
	if err != nil {
		return err
	}
	err = ugc.DisconnectUserGameLocation(userGameLocation)
	if err != nil {
		return err
	}
	return nil
}

func (ugc *UserGameLocationUsecase) LeaveInAudio(userGameLocationUsecase *model.UserGameLocation, roomID uint) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserGameLocations {
		if otherClient.UserID != userGameLocationUsecase.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "leave-audio",
				"roomID":     roomID,
				"fromUserID": userGameLocationUsecase.UserID,
				"toUserID":   otherClient.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := ugc.SendMessageToSpecificUser(userGameLocationUsecase, msg, otherClient.UserID)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				return err
			}

		}
	}
	return nil
}

func (ugc *UserGameLocationUsecase) DisconnectInGame(userGameLocation *model.UserGameLocation, roomID uint) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserGameLocations {
		if otherClient.UserID != userGameLocation.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "disconnect-game",
				"roomID":     roomID,
				"fromUserID": userGameLocation.UserID,
				"toUserID":   otherClient.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := ugc.SendMessageToSpecificUser(userGameLocation, msg, otherClient.UserID)
			if err != nil {
				return err
			}
		}
	}
	err := ugc.userGameLocationRepo.UpdateUserGameLocation(userGameLocation)
	if err != nil {
		return err
	}
	err = ugc.DisconnectUserGameLocation(userGameLocation)
	if err != nil {

		return err
	}
	return nil
}

func (ugc *UserGameLocationUsecase) DisconnectInAudio(userGameLocation *model.UserGameLocation, roomID uint) error {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(roomID)
	for _, otherClient := range connectedUserGameLocations {
		if otherClient.UserID != userGameLocation.UserID {
			leaveMsg := map[string]interface{}{
				"type":       "disconnect-audio",
				"roomID":     roomID,
				"fromUserID": userGameLocation.UserID,
				"toUserID":   otherClient.UserID,
			}
			msg := model.NewMessage(leaveMsg)
			err := ugc.SendMessageToSpecificUser(userGameLocation, msg, otherClient.UserID)
			if err != nil {
				return err
			}
		}
	}
	err := ugc.userGameLocationRepo.UpdateUserGameLocation(userGameLocation)
	if err != nil {
		return err
	}
	err = ugc.DisconnectUserGameLocation(userGameLocation)
	if err != nil {

		return err
	}
	return nil
}

func (ugc *UserGameLocationUsecase) GetSerializedConnectedUserGameLocations(roomID uint) ([]map[string]interface{}, error) {
	connectedUserGameLocations := ugc.inMemoryUserGameLocationRepo.GetAllUserGameLocationsByRoomId(roomID)
	userGameLocations := []map[string]interface{}{}
	for _, otherUserGameLocation := range connectedUserGameLocations {
		userGameLocation, exists, err := ugc.userGameLocationRepo.GetUserGameLocation(otherUserGameLocation.UserID)

		if err != nil {
			ugc.DisconnectUserGameLocation(userGameLocation)
			return nil, err
		}
		if !exists {
			log.Printf("user game location does not exist for user ID: %d", otherUserGameLocation.UserID)
			err := ugc.userGameLocationRepo.AddUserGameLocation(userGameLocation)
			if err != nil {
				ugc.DisconnectUserGameLocation(userGameLocation)
				return nil, fmt.Errorf("failed to add user location: %w", err)
			}
		}
		userGameLocationJSON, err := userGameLocation.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user location from JSON: %w", err)

		}
		var userGameLocationMap map[string]interface{}
		err = json.Unmarshal(userGameLocationJSON, &userGameLocationMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user location from JSON: %w", err)
		}
		userGameLocations = append(userGameLocations, userGameLocationMap)
	}
	return userGameLocations, nil
}

func (ugc *UserGameLocationUsecase) PingUserGameLocation(userGameLocation *model.UserGameLocation) error {
	userGameLocations, err := ugc.GetSerializedConnectedUserGameLocations(userGameLocation.RoomID)
	if err != nil {
		return err
	}
	pongMsg := map[string]interface{}{
		"type":              "pong",
		"fromUserID":        userGameLocation.UserID,
		"roomID":            userGameLocation.RoomID,
		"userGameLocations": userGameLocations,
	}
	msg := model.NewMessage(pongMsg)
	err = ugc.SendMessageToSameRoom(userGameLocation, msg)
	if err != nil {
		return err
	}
	return nil
}
