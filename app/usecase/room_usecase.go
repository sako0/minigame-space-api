package usecase

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type RoomUsecase struct {
	roomRepo repository.RoomRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo}
}

var ErrRoomNotFound = errors.New("room not found")

func (uc *RoomUsecase) JoinRoom(roomID uint, userID string, conn *websocket.Conn) (*model.Room, error) {
	room, ok := uc.roomRepo.LoadRoom(roomID)
	if !ok {
		room = &model.Room{}
	}

	player := &model.Player{UserID: userID, Conn: conn}
	room.Players = append(room.Players, player)

	uc.roomRepo.StoreRoom(room)

	return room, nil
}

func (uc *RoomUsecase) RemovePlayer(roomID uint, targetPlayerID uint) error {
	room, ok := uc.roomRepo.LoadRoom(roomID)
	if !ok {
		return ErrRoomNotFound
	}

	for i, player := range room.Players {
		if player.ID == targetPlayerID {
			room.Players = append(room.Players[:i], room.Players[i+1:]...)
			break
		}
	}

	uc.roomRepo.StoreRoom(room)

	return nil
}

func (uc *RoomUsecase) GetRoom(roomID uint) (*model.Room, error) {
	room, ok := uc.roomRepo.LoadRoom(roomID)
	if !ok {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

func (uc *RoomUsecase) ListRooms() ([]*model.Room, error) {
	rooms := uc.roomRepo.ListRooms()

	var roomsInfo []*model.Room
	for _, room := range rooms {
		Room := &model.Room{
			Name: room.Name,
		}
		for _, player := range room.Players {
			Room.Players = append(Room.Players, &model.Player{
				UserID: player.UserID,
			})
		}
		roomsInfo = append(roomsInfo, Room)
	}

	return roomsInfo, nil
}

func (uc *RoomUsecase) Broadcast(room *model.Room, msg *model.WsMessage) {
	for _, player := range room.Players {
		if player.Conn != nil {
			player.Conn.WriteJSON(msg)
		}
	}
}

func (uc *RoomUsecase) LeaveRoom(roomID uint, playerID uint) error {
	room, ok := uc.roomRepo.LoadRoom(roomID)
	if !ok {
		return ErrRoomNotFound
	}

	for i, player := range room.Players {
		if player.ID == playerID {
			room.Players = append(room.Players[:i], room.Players[i+1:]...)
			break
		}
	}

	uc.roomRepo.StoreRoom(room)

	return nil
}
