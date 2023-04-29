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
		userLocation = model.NewUserLocation(user, room, area, conn)

		h.roomUsecase.ConnectUserLocation(room, user, conn)

		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				connectedUserIds, err := h.roomUsecase.SendRoomJoinedEvent(userLocation)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)
					break
				}
				log.Println(connectedUserIds)
				roomJoinedMsg := map[string]interface{}{
					"type":             "client-joined",
					"connectedUserIds": connectedUserIds,
					"userId":           userLocation.UserID,
				}

				err = userLocation.Conn.WriteJSON(roomJoinedMsg)
				if err != nil {
					log.Printf("Error sending client-joined event to client: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)
				}

			case "leave-room":
				h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)

			case "offer", "answer", "ice-candidate":
				toUserIdStr, ok := msg["toUserId"].(string)
				if !ok {
					log.Printf("targetUserId is missing")
					return
				}
				toUserId, err := strconv.ParseUint(toUserIdStr, 10, 32)
				if err != nil {
					log.Printf("Invalid targetUserId: %v", err)
					return
				}

				toUserLocation, err := h.roomUsecase.FindUserLocationInRoom(uint(roomId), uint(toUserId))
				if err != nil {
					log.Printf("Error finding target user location: %v", err)
					return
				}

				err = toUserLocation.Conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending %s event to target user: %v", val.(string), err)
					h.roomUsecase.DisconnectUserLocation(userLocation.RoomID, userLocation)
					return
				}
			}
		}
	}
}
