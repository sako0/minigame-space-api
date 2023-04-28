package handler

import (
	"fmt"
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
			h.roomUsecase.DisconnectClient(userLocation.RoomID, userLocation)
			break
		}

		roomIdInt, err := strconv.Atoi(msg["roomId"].(string))
		if err != nil {
			log.Printf("roomId is missing")
			return
		}

		roomId := uint(roomIdInt)
		firebaseUid, ok := msg["userId"].(string)
		if !ok {
			log.Printf("firebaseUid is missing")
			return
		}
		areaId := uint(1)

		userLocation, err = h.roomUsecase.ConnectClient(roomId, areaId, firebaseUid, conn)
		if err != nil {
			log.Printf("Error connecting client: %v", err)
			break
		}
		log.Printf("userLocation: %v", userLocation)
		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				connectedUserIds, err := h.roomUsecase.SendRoomJoinedEvent(roomId)
				fmt.Println(connectedUserIds)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectClient(userLocation.RoomID, userLocation)
					break
				}

				roomJoinedMsg := map[string]interface{}{
					"type":             "client-joined",
					"connectedUserIds": connectedUserIds,
					"userId":           userLocation.User.ID,
				}

				err = userLocation.User.Conn.WriteJSON(roomJoinedMsg)
				if err != nil {
					log.Printf("Error sending client-joined event to client: %v", err)
					h.roomUsecase.DisconnectClient(userLocation.RoomID, userLocation)
				}
			case "leave-room":
				h.roomUsecase.DisconnectClient(userLocation.RoomID, userLocation)

				// 全てのクライアントに leave-room イベントをブロードキャスト
				leaveRoomMsg := map[string]interface{}{
					"type":   "leave-room",
					"userId": userLocation.User.ID,
				}

				err = h.roomUsecase.BroadcastMessageToOtherClients(userLocation, &model.Message{Payload: leaveRoomMsg})
				if err != nil {
					log.Printf("Error broadcasting leave-room event: %v", err)
				}

			case "offer", "answer", "ice-candidate":
				toUserIdFloat, ok := msg["toUserId"].(float64)
				if !ok {
					log.Printf("toUserId is missing")
					return
				}
				toUserId := uint(toUserIdFloat)
				// 送信先が自分自身でなければメッセージを送信する
				if toUserId != userLocation.User.ID {
					msgPayload := &model.Message{Payload: msg}
					h.roomUsecase.SendMessageToOtherClients(userLocation, toUserId, msgPayload)
				}
			}
		}
	}
}
