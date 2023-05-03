package handler

import (
	"log"
	"net/http"

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
	if conn == nil {
		log.Printf("conn is nil")
		return
	}

	defer conn.Close()

	userLocation := &model.UserLocation{}

	for {
		if userLocation.ID == 0 {
			log.Printf("Warning: userLocation is nil")
		}
		var msg = map[string]interface{}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			if userLocation.ID == 0 {
				h.roomUsecase.DisconnectUserLocation(userLocation)
			}
			return
		}

		firebaseUid, ok := msg["fromFirebaseUid"].(string)
		if !ok {
			log.Printf("firebaseUid is missing")
			return
		}

		user, err := h.roomUsecase.GetUserByFirebaseUID(firebaseUid)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return
		}
		log.Printf("Retrieved user ID: %d", user.ID)

		log.Printf("msg: %v", msg)

		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				if userLocation.ID != 0 {
					log.Printf("Warning: join-room userLocation is not nil")
				}
				roomIdFloat, ok := msg["roomId"].(float64)
				if !ok {
					log.Printf("roomIdFloat cast error")
				}
				roomId := uint(int(roomIdFloat))
				log.Printf("roomId: %d", roomId)

				room, err := h.roomUsecase.GetRoom(roomId)
				if err != nil {
					log.Printf("Error getting room: %v", err)
				}
				area, err := h.roomUsecase.GetArea(room.Area.ID)
				if err != nil {
					log.Printf("Error getting area: %v", err)
				}

				userLocation = &model.UserLocation{
					Area: area,
					Room: room,
					User: user,
					Conn: conn,
				}

				userLocation, err := h.roomUsecase.GetUserLocationByUserID(user.ID)
				if err != nil {
					log.Printf("Error checking user location existence: %v", err)
					return
				}

				if userLocation.ID == 0 {
					log.Println("handlerユーザーロケーションIDがない")
					return
				}
				// ユーザーの接続情報を作成し、userLocationに代入する
				userLocation, err = h.roomUsecase.ConnectUserLocation(userLocation)
				if err != nil {
					log.Printf("Error connecting user location: %v", err)
					return
				}
				// // 接続情報にconnを設定する
				userLocation.Conn = conn

				// Roomにjoinしたことを送信する
				connectedUserIds, err := h.roomUsecase.SendRoomJoinedEvent(userLocation)
				if err != nil {
					log.Printf("Error sending room joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation)
					return
				}

				// 自分自身にもclient-joinedイベントを送信する
				clientJoinedMsg := map[string]interface{}{
					"type":             "client-joined",
					"fromFirebaseUid":  user.FirebaseUID,
					"connectedUserIds": connectedUserIds,
				}

				if err := userLocation.Conn.WriteJSON(clientJoinedMsg); err != nil {
					log.Printf("Error sending client-joined event: %v", err)
					h.roomUsecase.DisconnectUserLocation(userLocation)
					return
				}

			case "leave-room":
				h.connectionUsecase.RemoveConnection(userLocation)
				h.roomUsecase.DisconnectUserLocation(userLocation)

			case "offer", "answer", "ice-candidate":
				toFirebaseUid := msg["toFirebaseUid"].(string)
				fromFirebaseUid := msg["fromFirebaseUid"].(string)
				if toFirebaseUid == fromFirebaseUid {
					log.Printf("Warning: toFirebaseUid is same as fromFirebaseUid")
					return
				}
				log.Printf("msg: %v", msg)
				log.Printf("toFirebaseUid: %v", toFirebaseUid)
				toTargetConn, ok := h.connectionUsecase.GetConnectionByUserFirebaseUID(toFirebaseUid)
				if !ok {
					log.Printf("Info getting target connection: %v", err)
					return
				}
				if toTargetConn != nil {
					log.Printf("ユーザーID:[ %s ]から [ %v ] に送られました。メッセージタイプは[ %v ]です。", userLocation.User.FirebaseUID, toFirebaseUid, msg["type"])
					h.roomUsecase.HandleSignalMessage(userLocation, toTargetConn, msg)
				}
			}
		}
	}
}
