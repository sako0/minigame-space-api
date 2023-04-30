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
	roomUsecase       usecase.RoomUsecase
	connectionUsecase usecase.ConnectionStoreUsecase
	upgrader          websocket.Upgrader
}

func NewWebSocketHandler(roomUsecase usecase.RoomUsecase, connectionUsecase usecase.ConnectionStoreUsecase, upgrader websocket.Upgrader) *WebSocketHandler {
	return &WebSocketHandler{
		roomUsecase:       roomUsecase,
		connectionUsecase: connectionUsecase,
		upgrader:          upgrader,
	}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	var user *model.User
	var userLocation *model.UserLocation

	// Handle incoming messages
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

		log.Printf("msg: %v", msg)

		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
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

				// ユーザーの接続情報を作成し、userLocationに代入する
				userLocation = &model.UserLocation{
					User: *user,
					Area: *area,
					Room: *room,
					Conn: conn,
				}

				// userLocationを接続する
				userLocation, err = h.roomUsecase.ConnectUserLocation(userLocation)
				if err != nil {
					log.Printf("Error connecting user location: %v", err)
					return
				}

				// 接続情報にconnを設定する
				userLocation.Conn = conn
				// ConnectionStoreに接続を保存する
				h.connectionUsecase.StoreUserLocation(userLocation)

				// Roomにjoinしたことを送信する
				_, err = h.roomUsecase.SendRoomJoinedEvent(userLocation)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation)
					break
				}

			case "leave-room":
				h.connectionUsecase.RemoveConnection(user)
				h.roomUsecase.DisconnectUserLocation(userLocation)

			case "offer", "answer", "ice-candidate":
				toUserId := uint(msg["toUserId"].(float64))
				targetConn, ok := h.connectionUsecase.GetConnectionByUserID(toUserId)
				if !ok {
					log.Printf("Info getting target connection: %v", err)
					return
				}
				if targetConn != nil {
					log.Println(targetConn)
					// Process offer, answer, ice-candidate events
					log.Printf("ユーザーID:[ %d ]から [ %d ] に送られました。メッセージタイプは[ %v ]です。", userLocation.User.ID, toUserId, msg["type"])

					h.roomUsecase.HandleSignalMessage(userLocation, targetConn, msg)
				}
			}
		}
	}
}
