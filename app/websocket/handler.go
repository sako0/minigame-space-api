package handler

import (
	"log"
	"net/http"

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

	client := &model.Client{}

	for {
		var msg = map[string]interface{}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			h.roomUsecase.DisconnectClient(client.RoomId(), client)
			break
		}

		roomIdFloat, ok := msg["roomId"].(float64)
		if !ok {
			log.Printf("roomIdFloat cast error")
		}
		roomId := uint(int(roomIdFloat))
		if !ok {
			log.Printf("roomId is missing")
			return
		}
		userId, ok := msg["fromFirebaseUid"].(string)
		if !ok {
			log.Printf("fromFirebaseUid is missing")
			return
		}
		client = model.NewClient(conn, roomId, userId)

		h.roomUsecase.ConnectClient(roomId, userId, conn)

		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				connectedUserIds, err := h.roomUsecase.SendRoomJoinedEvent(client)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectClient(client.RoomId(), client)
					break
				}

				roomJoinedMsg := map[string]interface{}{
					"type":             "client-joined",
					"connectedUserIds": connectedUserIds,
					"userId":           client.UserId(),
				}

				err = client.Conn().WriteJSON(roomJoinedMsg)
				if err != nil {
					log.Printf("Error sending client-joined event to client: %v", err)
					h.roomUsecase.DisconnectClient(client.RoomId(), client)
				}
			case "leave-room":
				h.roomUsecase.DisconnectClient(client.RoomId(), client)

				// 全てのクライアントに leave-room イベントをブロードキャスト
				leaveRoomMsg := map[string]interface{}{
					"type":            "leave-room",
					"fromFirebaseUid": client.UserId(),
				}

				err = h.roomUsecase.BroadcastMessageToOtherClients(client, &model.Message{Payload: leaveRoomMsg})
				if err != nil {
					log.Printf("Error broadcasting leave-room event: %v", err)
				}

			case "offer", "answer", "ice-candidate":
				toUserId, ok := msg["toFirebaseUid"].(string)
				if !ok {
					log.Printf("toFirebaseUid is missing")
					return
				}
				// 送信先が自分自身でなければメッセージを送信する
				if toUserId != client.UserId() {
					msgPayload := &model.Message{Payload: msg}
					h.roomUsecase.SendMessageToOtherClients(client, toUserId, msgPayload)
				}
			}
		}
	}
}
