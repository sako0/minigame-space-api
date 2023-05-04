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
	roomUsecase usecase.RoomUsecase
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(roomUsecase usecase.RoomUsecase, upgrader websocket.Upgrader) *WebSocketHandler {
	return &WebSocketHandler{roomUsecase: roomUsecase, upgrader: upgrader}
}

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	client := &model.Client{
		Conn: conn,
	}

	for {
		msg, err := h.readMessage(conn)
		if err != nil {
			h.handleDisconnect(client)
			log.Printf("Error reading message: %v", err)
			break
		}

		client, err = h.processMessage(client, msg)
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

func (h *WebSocketHandler) processMessage(client *model.Client, msg map[string]interface{}) (*model.Client, error) {
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

func (h *WebSocketHandler) handleJoinRoom(client *model.Client, msg map[string]interface{}) (*model.Client, error) {
	roomId := uint(msg["roomId"].(float64))
	if roomId == 0 {
		return client, fmt.Errorf("invalid roomId")
	}
	client.RoomId = roomId

	fromFirebaseUId, ok := msg["fromFirebaseUid"].(string)
	if !ok {
		return client, fmt.Errorf("invalid fromFirebaseUid")
	}
	client.UserId = fromFirebaseUId

	clientPt, err := h.roomUsecase.ConnectClient(*client)
	if err != nil {
		log.Printf("Error connecting client to room: %v", err)
		return client, err
	}
	client = clientPt
	connectedUserIds, err := h.roomUsecase.SendRoomJoinedEvent(client)
	if err != nil {
		log.Printf("Error sending room joined event: %v", err)
		h.roomUsecase.DisconnectClient(client.Room.ID, client)
		return nil, err
	}
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"userId":           client.UserId,
	}
	log.Printf("client Conn: %v", client.Conn)
	err = client.Conn.WriteJSON(roomJoinedMsg)
	if err != nil {
		log.Printf("Error sending client-joined event to client: %v", err)
		h.roomUsecase.DisconnectClient(client.Room.ID, client)
	}
	return client, nil
}

func (h *WebSocketHandler) handleLeaveRoom(client *model.Client, msg map[string]interface{}) (*model.Client, error) {
	h.roomUsecase.DisconnectClient(client.RoomId, client)
	// 全てのクライアントに leave-room イベントをブロードキャスト
	leaveRoomMsg := map[string]interface{}{
		"type":            "leave-room",
		"fromFirebaseUid": client.UserId,
	}

	err := h.roomUsecase.BroadcastMessageToOtherClients(client, &model.Message{Payload: leaveRoomMsg})
	if err != nil {
		log.Printf("Error broadcasting leave-room event: %v", err)
		return nil, err
	}
	return client, nil
}

func (h *WebSocketHandler) handleSignalingMessage(client *model.Client, msg map[string]interface{}) (*model.Client, error) {
	toFirebaseUid, ok := msg["toFirebaseUid"].(string)
	if !ok {
		log.Printf("toFirebaseUid is missing")
		return nil, fmt.Errorf("toFirebaseUid is missing")
	}
	// 送信先が自分自身でなければメッセージを送信する
	if toFirebaseUid != client.UserId {
		// 来たメッセージをそのまま送信する
		msgPayload := &model.Message{Payload: msg}
		h.roomUsecase.SendMessageToOtherClients(client, toFirebaseUid, msgPayload)
	}
	return client, nil
}

func (h *WebSocketHandler) handleDisconnect(client *model.Client) error {
	h.roomUsecase.DisconnectClient(client.RoomId, client)
	leaveRoomMsg := map[string]interface{}{
		"type":            "leave-room",
		"fromFirebaseUid": client.UserId,
	}
	err := h.roomUsecase.BroadcastMessageToOtherClients(client, &model.Message{Payload: leaveRoomMsg})
	if err != nil {
		log.Printf("Error broadcasting leave-room event: %v", err)
		return err
	}
	return nil
}
