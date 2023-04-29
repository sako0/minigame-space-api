package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/usecase"
)

type WebSocketHandler struct {
	roomUsecase usecase.RoomUsecase
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(roomUsecase usecase.RoomUsecase, upgrader websocket.Upgrader) *WebSocketHandler {
	return &WebSocketHandler{roomUsecase: roomUsecase, upgrader: upgrader}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userLocation := &model.UserLocation{}

	for {
		var msg = map[string]interface{}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)
			break
		}

		roomIdStr, ok := msg["roomId"].(string)
		if !ok {
			log.Printf("roomId is missing")
			return
		}
		roomId, err := strconv.ParseUint(roomIdStr, 10, 32)
		if err != nil {
			log.Printf("Invalid roomId: %v", err)
			return
		}
		firebaseUid, ok := msg["userId"].(string)
		if !ok {
			log.Printf("userId is missing")
			return
		}

		user, err := h.roomUsecase.GetUserByFirebaseUID(firebaseUid)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return
		}

		room, err := h.roomUsecase.GetRoom(uint(roomId))
		if err != nil {
			log.Printf("Error getting room: %v", err)
			return
		}

		area, err := h.roomUsecase.GetArea(room.AreaID)
		if err != nil {
			log.Printf("Error getting area: %v", err)
			return
		}
		userLocation = model.NewUserLocation(user, area, room, 1, 1, 1, conn)

		userLocation, err = h.roomUsecase.ConnectUserLocation(userLocation)
		if err != nil {
			log.Printf("Error connecting user location: %v", err)
			return
		}

		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				_, err := h.roomUsecase.SendRoomJoinedEvent(userLocation)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)
					break
				}

			case "leave-room":
				h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)

			case "offer", "answer", "ice-candidate":
				toUserId, ok := msg["toUserId"].(float64)
				if !ok {
					log.Printf("toUserId is missing")
					return
				}

				toUserLocations, err := h.roomUsecase.FindUserLocationInRoom(room, uint(toUserId))
				if err != nil {
					log.Printf("Error finding target user location: %v", err)
					return
				}

				for _, toUserLocation := range toUserLocations {
					if toUserLocation.Conn == nil {
						err := toUserLocation.Conn.WriteJSON(msg)
						if err != nil {
							log.Printf("Error writing JSON: %v", err)
							return
						}
					}
				}
			}
		}
	}
}
