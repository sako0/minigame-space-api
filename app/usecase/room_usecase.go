package usecase

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type RoomUsecase struct {
	roomRepo repository.RoomRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo}
}

func (uc *RoomUsecase) ConnectClient(roomId string, userId string, conn *websocket.Conn) (*model.Client, error) {

	client := model.NewClient(conn, roomId, userId)

	room, ok := uc.roomRepo.GetRoom(roomId)
	if !ok {
		room = model.NewRoom()
		uc.roomRepo.AddRoom(roomId, room)
	}
	// 既に同じクライアントが接続している場合は何もしない
	for otherClient := range room.Clients {
		if otherClient.RoomId() == roomId && otherClient.UserId() == userId {
			return nil, nil
		}
	}
	room.Clients[client] = true

	return client, nil
}

func (uc *RoomUsecase) DisconnectClient(roomId string, client *model.Client) {
	room, ok := uc.roomRepo.GetRoom(roomId)
	if ok {
		delete(room.Clients, client)
		if len(room.Clients) == 0 {
			uc.roomRepo.RemoveRoom(roomId)
		}
	}
	if client.Conn() != nil {
		client.Conn().Close()
	}
}

func (uc *RoomUsecase) SendRoomJoinedEvent(client *model.Client) ([]string, error) {
	roomId := client.RoomId()

	room, ok := uc.roomRepo.GetRoom(roomId)
	if !ok {
		return nil, fmt.Errorf("Room not found: %s", roomId)
	}
	connectedUserIds := []string{}
	for otherClient := range room.Clients {
		if otherClient != client {
			connectedUserIds = append(connectedUserIds, otherClient.UserId())
		}
	}

	return connectedUserIds, nil
}

func (uc *RoomUsecase) SendMessageToOtherClients(client *model.Client, toUserId string, msg *model.Message) {
	room, ok := uc.roomRepo.GetRoom(client.RoomId())
	if !ok {
		log.Printf("Room not found: %s", client.RoomId())
		return
	}

	msgPayload := msg.Payload
	msgPayload["userId"] = client.UserId()

	for otherClient := range room.Clients {
		if otherClient.UserId() == toUserId {
			err := otherClient.Conn().WriteJSON(msgPayload)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				uc.DisconnectClient(client.RoomId(), otherClient)
			}
		}
	}
}
