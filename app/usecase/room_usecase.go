package usecase

import (
	"log"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type RoomUsecase struct {
	roomRepo   repository.RoomRepository
	clientRepo repository.ClientRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository, clientRepo repository.ClientRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo, clientRepo: clientRepo}
}

func (uc *RoomUsecase) ConnectClient(client model.Client) (*model.Client, error) {
	room, ok := uc.roomRepo.GetRoom(client.RoomId)
	if !ok {
		// ルームが存在しない場合は新規作成
		room = model.NewRoom(client.RoomId)
		uc.roomRepo.AddRoom(client.RoomId, room)
	}

	// 既に同じクライアントが接続している場合は何もしない
	for _, otherClient := range uc.clientRepo.GetAllClientsByRoomId(room.ID) {
		if otherClient.RoomId == room.ID && otherClient.UserId == client.UserId {
			return &client, nil
		}
	}

	uc.clientRepo.AddClient(&client)

	return &client, nil
}

func (uc *RoomUsecase) DisconnectClient(roomId uint, client *model.Client) {
	_, ok := uc.roomRepo.GetRoom(roomId)
	if ok {
		uc.clientRepo.RemoveClient(client.UserId)
		if len(uc.clientRepo.GetAllClientsByRoomId(roomId)) == 0 {
			uc.roomRepo.RemoveRoom(roomId)
		}
	}
	if client.Conn != nil {
		client.Conn.Close()
	}
}

func (uc *RoomUsecase) BroadcastMessageToOtherClients(client *model.Client, msg *model.Message) error {
	roomId := client.RoomId

	msgPayload := msg.Payload
	msgPayload["fromFirebaseUid"] = client.UserId

	for _, otherClient := range uc.clientRepo.GetAllClientsByRoomId(roomId) {
		if otherClient != client {
			err := otherClient.Conn.WriteJSON(msgPayload)
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
				uc.DisconnectClient(roomId, otherClient)
			}
		}
	}

	return nil
}

func (uc *RoomUsecase) SendRoomJoinedEvent(client *model.Client) ([]string, error) {
	roomId := client.RoomId
	connectedUserIds := []string{}
	for _, otherClient := range uc.clientRepo.GetAllClientsByRoomId(roomId) {
		if otherClient != client {
			connectedUserIds = append(connectedUserIds, otherClient.UserId)
		}
	}

	return connectedUserIds, nil
}

func (uc *RoomUsecase) SendMessageToOtherClients(client *model.Client, toUserId string, msg *model.Message) {
	room, ok := uc.roomRepo.GetRoom(client.RoomId)
	if !ok {
		log.Printf("Room not found: %d", client.RoomId)
		return
	}

	msgPayload := msg.Payload
	msgPayload["fromFirebaseUid"] = client.UserId

	for _, otherClient := range uc.clientRepo.GetAllClientsByRoomId(room.ID) {
		if otherClient.UserId == toUserId {
			err := otherClient.Conn.WriteJSON(msgPayload)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				uc.DisconnectClient(client.RoomId, otherClient)
			}
		}
	}
}

func (uc *RoomUsecase) GetRoom(roomId uint) (*model.Room, bool) {
	return uc.roomRepo.GetRoom(roomId)
}
