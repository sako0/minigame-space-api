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
	return &WebSocketHandler{
		roomUsecase: roomUsecase,
		upgrader:    upgrader,
	}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// defer conn.Close()

	var user *model.User
	var userLocation *model.UserLocation

	for {
		var msg = map[string]interface{}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			if userLocation != nil {
				h.roomUsecase.DisconnectUserLocation(userLocation)
			}
			break
		}

		if user == nil {
			firebaseUid, ok := msg["userId"].(string)
			if !ok {
				log.Printf("userId is missing")
				return
			}

			user, err = h.roomUsecase.GetUserByFirebaseUID(firebaseUid)
			if err != nil {
				log.Printf("Error getting user: %v", err)
				return
			}
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

				h.roomUsecase.StoreConnection(user, conn)
				log.Printf("connectionsがこのメモリにstoreされました: %v", conn)
				_, err = h.roomUsecase.SendRoomJoinedEvent(userLocation)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation)
					break
				}

			case "leave-room":
				h.roomUsecase.RemoveConnection(user)
				h.roomUsecase.DisconnectUserLocation(userLocation)
			case "offer", "answer", "ice-candidate":
				toUserId, _ := msg["toUserId"].(float64)
				if toUserId == 0 {
					log.Printf("toUserId is missing")
					return
				}

				toUserLocation, err := h.roomUsecase.FindUserLocationInRoom(uint(toUserId))
				if err != nil {
					log.Printf("Error finding user location in room: %v", err)
					return
				}
				if toUserLocation == nil {
					log.Printf("toUserLocation is missing")
					return
				}
				// Check if WebSocket connection exists for the target user
				toUserConn, ok := h.roomUsecase.GetConnectionByUserID(toUserLocation.UserID)
				if !ok {
					log.Printf("Info: websocket.Conn not found for user %d", toUserLocation.UserID)
					continue // Skip sending the message and continue with the next iteration
				}
				err = toUserConn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error writing JSON: %v", err)
					return
				}

			}
		}
	}
}
