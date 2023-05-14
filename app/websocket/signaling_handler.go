package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

// websocketの接続を維持するためのping/pongの間隔
var pongWait = 3600 * time.Second

func (h *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	// websocketの接続を維持するためのping/pongの間隔を設定
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	userLocation := model.NewUserLocationByConn(conn)

	defer func() {
		// クリーンアップ処理
		err := h.disconnect(userLocation)
		if err != nil {
			log.Printf("Error disconnecting user: %v", err)
		}
	}()

	for {
		msg, err := h.readMessage(conn)
		if err != nil {
			log.Printf("Error reading message msg: %v", msg)
			log.Printf("Error reading message: %v", err)
			break
		}

		err = h.processMessage(userLocation, msg)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			break
		}
	}
}

var retryInterval = 500 * time.Millisecond

func (h *WebSocketHandler) readMessage(conn *websocket.Conn) (map[string]interface{}, error) {
	var msg = map[string]interface{}{}
	err := conn.ReadJSON(&msg)
	if err != nil {
		log.Printf("error reading JSON msg: %v", msg)
		// 一時的なエラーの場合はリトライ
		if isTemporary(err) {
			time.Sleep(retryInterval)
			return h.readMessage(conn)
		}
		log.Printf("Error reading JSON: %v", err)
		return nil, err
	}
	return msg, nil
}

func (h *WebSocketHandler) processMessage(client *model.UserLocation, msg map[string]interface{}) error {
	var err error
	switch msg["type"].(string) {
	case "join-area":
		err = h.handleJoinArea(client, msg)
	case "join-room":
		err = h.handleJoinRoom(client, msg)
	case "leave-area":
		err = h.handleLeaveArea(client, msg)
	case "leave-room":
		err = h.handleLeaveRoom(client, msg)
	case "move":
		err = h.handleMove(client, msg)
	case "offer", "answer", "ice-candidate":
		err = h.handleSignalingMessage(client, msg)
	default:
		err = fmt.Errorf("unknown message type")
	}
	if err != nil {
		// 一時的なエラーの場合はリトライ
		if isTemporary(err) {
			time.Sleep(retryInterval)
			return h.processMessage(client, msg)
		}
		log.Printf("Error processing message: %v", err)
	}
	return nil
}

func (h *WebSocketHandler) handleJoinArea(userLocation *model.UserLocation, msg map[string]interface{}) error {
	areaID := uint(msg["areaID"].(float64))
	userLocation.AreaID = areaID

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	userLocation.UserID = fromUserID

	err := h.userLocationUsecase.ConnectUserLocationForArea(userLocation)
	if err != nil {
		log.Printf("Error connecting client to area: %v", err)
		return err
	}
	err = h.userLocationUsecase.SendAreaJoinedEvent(userLocation)
	if err != nil {
		log.Printf("Error sending area joined event: %v", err)
		h.userLocationUsecase.DisconnectUserLocation(userLocation)
		return err
	}

	return nil
}

func (h *WebSocketHandler) handleJoinRoom(userLocation *model.UserLocation, msg map[string]interface{}) error {
	roomId := uint(msg["roomID"].(float64))
	if !isValidRoomId(roomId) {
		return fmt.Errorf("invalid roomID")
	}
	userLocation.RoomID = roomId

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	userLocation.UserID = fromUserID

	err := h.userLocationUsecase.ConnectUserLocationForRoom(userLocation)
	if err != nil {
		log.Printf("Error connecting client to room: %v", err)
		return err
	}
	err = h.userLocationUsecase.SendRoomJoinedEvent(userLocation)
	if err != nil {
		log.Printf("Error sending room joined event: %v", err)
		h.userLocationUsecase.DisconnectUserLocation(userLocation)
		return err
	}

	return nil
}
func (h *WebSocketHandler) handleLeaveArea(userLocation *model.UserLocation, msg map[string]interface{}) error {
	return h.userLocationUsecase.LeaveInArea(userLocation)
}
func (h *WebSocketHandler) handleLeaveRoom(userLocation *model.UserLocation, msg map[string]interface{}) error {
	return h.userLocationUsecase.LeaveInRoom(userLocation)
}

func (h *WebSocketHandler) handleMove(userLocation *model.UserLocation, msg map[string]interface{}) error {
	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	areaID := uint(msg["areaID"].(float64))
	userLocation.UserID = fromUserID
	userLocation.AreaID = areaID
	xAxis := int(msg["xAxis"].(float64))
	yAxis := int(msg["yAxis"].(float64))

	err := h.userLocationUsecase.MoveInArea(userLocation, xAxis, yAxis)
	if err != nil {
		log.Printf("Error updating and broadcasting user location: %v", err)
		return err
	}

	return nil
}

func (h *WebSocketHandler) disconnect(userLocation *model.UserLocation) error {
	// TODO: 仮にLeaveInRoomをしているが、実際にはDisconnectInAllを使う
	return h.userLocationUsecase.LeaveInRoom(userLocation)
}

func (h *WebSocketHandler) handleSignalingMessage(userLocation *model.UserLocation, msg map[string]interface{}) error {
	toUserID := uint(msg["toUserID"].(float64))
	if !isValidUserId(toUserID) {
		return fmt.Errorf("invalid toUserID")
	}
	msgPayload := &model.Message{Payload: msg}
	// 特定のユーザーにメッセージを送信する(ここでルーム全員に送信するとブラウザ側でメモリエラーになる)
	h.userLocationUsecase.SendMessageToSpecificUser(userLocation, msgPayload, toUserID)
	return nil
}

// ヘルパー関数
func isValidRoomId(roomId uint) bool {
	return roomId != 0
}

func isValidUserId(fromUserID uint) bool {
	return fromUserID != 0
}

// 一時的なエラーかどうかを判定する
func isTemporary(err error) bool {
	te, ok := err.(interface {
		Temporary() bool
	})
	return ok && te.Temporary()
}
