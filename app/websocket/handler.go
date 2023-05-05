package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/usecase"
)

type WebSocketHandler struct {
	userLocationUsecase usecase.UserLocationUsecase
	upgrader            websocket.Upgrader
}

func NewWebSocketHandler(userLocationUsecase usecase.UserLocationUsecase, upgrader websocket.Upgrader) *WebSocketHandler {
	return &WebSocketHandler{userLocationUsecase: userLocationUsecase, upgrader: upgrader}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	userLocation := model.NewUserLocationByConn(conn)

	for {
		msg, err := h.readMessage(conn)
		if err != nil {
			h.handleDisconnect(userLocation)
			log.Printf("Error reading message: %v", err)
			break
		}

		userLocation, err = h.processMessage(userLocation, msg)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			break
		}
	}
}

func (h *WebSocketHandler) readMessage(conn *websocket.Conn) (map[string]interface{}, error) {
	var msg = map[string]interface{}{}
	err := conn.ReadJSON(&msg)
	if err != nil {
		log.Printf("Error reading JSON: %v", err)
		return nil, err
	}
	return msg, nil
}

func (h *WebSocketHandler) processMessage(client *model.UserLocation, msg map[string]interface{}) (*model.UserLocation, error) {
	switch msg["type"].(string) {
	case "join-room":
		return h.handleJoinRoom(client, msg)
	case "leave-room":
		return h.handleLeaveRoom(client, msg)
	case "offer", "answer", "ice-candidate":
		return h.handleSignalingMessage(client, msg)
	default:
		return client, fmt.Errorf("unknown message type")
	}
}

func (h *WebSocketHandler) handleJoinRoom(userLocation *model.UserLocation, msg map[string]interface{}) (*model.UserLocation, error) {
	roomId := uint(msg["roomId"].(float64))
	if !isValidRoomId(roomId) {
		return userLocation, fmt.Errorf("invalid roomId")
	}
	userLocation.RoomID = roomId

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidFromFirebaseUid(fromUserID) {
		return userLocation, fmt.Errorf("invalid fromUserID")
	}
	userLocation.UserID = fromUserID

	userLocationPT, err := h.userLocationUsecase.ConnectUserLocation(userLocation)
	if err != nil {
		log.Printf("Error connecting client to room: %v", err)
		return userLocationPT, err
	}
	userLocation = userLocationPT
	connectedUserIds, err := h.userLocationUsecase.SendRoomJoinedEvent(userLocation)
	if err != nil {
		log.Printf("Error sending room joined event: %v", err)
		h.userLocationUsecase.DisconnectUserLocation(userLocation)
		return nil, err
	}
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"fromUserID":       userLocation.UserID,
	}
	err = userLocation.Conn.WriteJSON(roomJoinedMsg)
	if err != nil {
		log.Printf("Error sending client-joined event to client: %v", err)
		h.userLocationUsecase.DisconnectUserLocation(userLocation)
	}
	return userLocation, nil
}

func (h *WebSocketHandler) handleLeaveRoom(userLocation *model.UserLocation, msg map[string]interface{}) (*model.UserLocation, error) {
	h.userLocationUsecase.DisconnectUserLocation(userLocation)
	// 全てのクライアントに leave-room イベントをブロードキャスト
	leaveRoomMsg := map[string]interface{}{
		"type":            "leave-room",
		"fromFirebaseUid": userLocation.UserID,
	}

	err := h.userLocationUsecase.BroadcastMessageToOtherClients(userLocation, &model.Message{Payload: leaveRoomMsg})
	if err != nil {
		log.Printf("Error broadcasting leave-room event: %v", err)
		return nil, err
	}
	return userLocation, nil
}

func (h *WebSocketHandler) handleSignalingMessage(userLocation *model.UserLocation, msg map[string]interface{}) (*model.UserLocation, error) {
	toUserID := uint(msg["toUserID"].(float64))
	if !isValidRoomId(toUserID) {
		return userLocation, fmt.Errorf("invalid toUserID")
	}

	// 送信先が自分自身でなければメッセージを送信する
	if toUserID != userLocation.UserID {
		// 来たメッセージをそのまま送信する
		msgPayload := &model.Message{Payload: msg}
		h.userLocationUsecase.SendMessageToOtherClients(userLocation, msgPayload)
	}
	return userLocation, nil
}

func (h *WebSocketHandler) handleDisconnect(userLocation *model.UserLocation) error {
	h.userLocationUsecase.DisconnectUserLocation(userLocation)
	leaveRoomMsg := map[string]interface{}{
		"type":       "leave-room",
		"fromUserID": userLocation.UserID,
	}
	err := h.userLocationUsecase.BroadcastMessageToOtherClients(userLocation, &model.Message{Payload: leaveRoomMsg})
	if err != nil {
		log.Printf("Error broadcasting leave-room event: %v", err)
		return err
	}
	return nil
}

func isValidRoomId(roomId uint) bool {
	return roomId != 0
}

func isValidFromFirebaseUid(fromUserID uint) bool {
	return fromUserID != 0
}
