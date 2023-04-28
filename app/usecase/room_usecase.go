package usecase

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type RoomUsecase struct {
	roomRepo         repository.RoomRepository
	areaRepo         repository.AreaRepository
	roomTypeRepo     repository.RoomTypeRepository
	userRepo         repository.UserRepository
	userLocationRepo repository.UserLocationRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository, areaRepo repository.AreaRepository, roomtypeRepo repository.RoomTypeRepository, userRepo repository.UserRepository, userLocationRepo repository.UserLocationRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo, areaRepo: areaRepo, roomTypeRepo: roomtypeRepo, userRepo: userRepo, userLocationRepo: userLocationRepo}
}

func (uc *RoomUsecase) ConnectClient(roomId uint, areaId uint, firebaseUid string, conn *websocket.Conn) (*model.UserLocation, error) {
	user, err := uc.userRepo.GetUser(firebaseUid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user = model.NewUser(conn, firebaseUid)
		err = uc.userRepo.AddUser(user)
		if err != nil {
			return nil, err
		}
	} else {
		err = uc.userRepo.UpdateUser(user)
		if err != nil {
			return nil, err
		}
	}
	room, err := uc.roomRepo.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	if room == nil {
		room = model.NewRoom(areaId, 1)
		err = uc.roomRepo.AddRoom(room)
		if err != nil {
			return nil, err
		}
	}

	area, err := uc.areaRepo.GetArea(areaId)
	if err != nil {
		return nil, err
	}
	if area == nil && err == nil {
		return nil, errors.New("area not found")
	}
	room.Area = *area

	roomType, err := uc.roomTypeRepo.GetRoomType(1)
	if err != nil {
		return nil, err
	}
	if roomType == nil {
		return nil, errors.New("room type not found")
	}
	room.RoomType = *roomType

	for _, otherUserLocation := range room.UserLocations {
		if otherUserLocation.RoomID == roomId && otherUserLocation.User.FirebaseUID == firebaseUid {
			return nil, nil
		}
	}
	userLocation := model.NewUserLocation(user, room, area)
	log.Println(userLocation)
	err = uc.userLocationRepo.SaveUserLocation(userLocation)
	if err != nil {
		return nil, err
	}

	if userLocation.User == (model.User{}) {
		userLocation.User = *user
	} else {
		userLocation.User.Conn = user.Conn
		userLocation.User.FirebaseUID = user.FirebaseUID
	}

	room.UserLocations = append(room.UserLocations, *userLocation)

	return userLocation, nil
}

func (uc *RoomUsecase) DisconnectClient(roomId uint, userLocation *model.UserLocation) {
	room, err := uc.roomRepo.GetRoom(roomId)
	if err != nil {
		return
	}
	if room != nil {
		for i, otherUserLocation := range room.UserLocations {
			if &otherUserLocation == userLocation {
				room.UserLocations = append(room.UserLocations[:i], room.UserLocations[i+1:]...)
				break
			}
		}
		if len(room.UserLocations) == 0 {
			uc.roomRepo.RemoveRoom(roomId)
		}
	}
	if userLocation.User != (model.User{}) && userLocation.User.Conn != nil {
		userLocation.User.Conn.Close()
	}
}

func (uc *RoomUsecase) BroadcastMessageToOtherClients(userLocation *model.UserLocation, msg *model.Message) error {
	roomId := userLocation.RoomID
	room, err := uc.roomRepo.GetRoom(roomId)
	if err != nil {
		return err
	}
	if room == nil {
		room = model.NewRoom(userLocation.AreaID, 1)
	}

	msgPayload := msg.Payload
	msgPayload["userId"] = userLocation.User.ID

	for _, otherUserLocation := range room.UserLocations {
		if &otherUserLocation != userLocation {
			if otherUserLocation.User.Conn != nil {
				err := otherUserLocation.User.Conn.WriteJSON(msgPayload)
				if err != nil {
					uc.DisconnectClient(roomId, &otherUserLocation)
				}
			}
		}
	}

	return nil
}

func (uc *RoomUsecase) SendRoomJoinedEvent(roomId uint) ([]uint, error) {
	room, err := uc.roomRepo.GetRoom(roomId)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room not found")
	}
	connectedUserIds := []uint{}
	for _, userLocation := range room.UserLocations {
		connectedUserIds = append(connectedUserIds, userLocation.User.ID)
	}

	return connectedUserIds, nil
}

func (uc *RoomUsecase) SendMessageToOtherClients(userLocation *model.UserLocation, toUserId uint, msg *model.Message) {
	room, err := uc.roomRepo.GetRoom(userLocation.RoomID)
	if err != nil {
		return
	}
	if room == nil {
		room = model.NewRoom(userLocation.AreaID, 1)
	}

	msgPayload := msg.Payload
	msgPayload["userId"] = userLocation.User.ID

	for _, otherUserLocation := range room.UserLocations {
		if otherUserLocation.User.ID == toUserId {
			err := otherUserLocation.User.Conn.WriteJSON(msgPayload)
			if err != nil {
				uc.DisconnectClient(userLocation.RoomID, &otherUserLocation)
			}
		}
	}
}

func (uc *RoomUsecase) CreateRoom(areaId uint, roomTypeId uint) (*model.Room, error) {
	area, err := uc.areaRepo.GetArea(areaId)
	if err != nil {
		return nil, err
	}
	if area == nil && err == nil {
		return nil, errors.New("area not found")
	}
	roomType, err := uc.roomTypeRepo.GetRoomType(roomTypeId)
	if err != nil {
		return nil, err
	}
	if roomType == nil {
		return nil, errors.New("room type not found")
	}

	room := model.NewRoom(areaId, roomTypeId)
	if err := uc.roomRepo.AddRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}
func (uc *RoomUsecase) JoinRoom(firebaseUid string, roomID uint) error {
	user, err := uc.userRepo.GetUser(firebaseUid)
	if err != nil {
		return err
	}
	if user == nil && err == nil {
		return errors.New("user not found")
	}

	room, err := uc.roomRepo.GetRoom(roomID)
	if err != nil {
		return err
	}
	if room == nil && err == nil {
		return errors.New("room not found")
	}

	userLocation := model.NewUserLocation(user, room, &room.Area)
	if err := uc.userLocationRepo.AddUserLocation(userLocation); err != nil {
		return err
	}

	return nil
}
