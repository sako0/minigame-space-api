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

type UserGameLocationHandler struct {
	userGameLocationUsecase usecase.UserGameLocationUsecase
	upgrader                websocket.Upgrader
}

func NewUserGameLocationHandler(userGameLocationUsecase usecase.UserGameLocationUsecase, upgrader websocket.Upgrader) *UserGameLocationHandler {
	return &UserGameLocationHandler{userGameLocationUsecase: userGameLocationUsecase, upgrader: upgrader}
}

const PingTimeout = 20 * time.Second

func (h *UserGameLocationHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	userGameLocation := model.NewUserGameLocationByConn(conn)

	defer func() {
		// クリーンアップ処理
		log.Println("disconnect userGameLocation:", userGameLocation.UserID)
		h.cleanUp(userGameLocation)

	}()

	// ゴルーチンを起動してピングの監視を行う
	lastPingTime := time.Now()
	go func() {
		for {
			time.Sleep(time.Second)
			if time.Since(lastPingTime) > PingTimeout {
				log.Printf("Ping timeout for user: %v", userGameLocation.UserID)
				h.cleanUp(userGameLocation)
				conn.Close()
				break
			}
		}
	}()
	for {
		msg, err := h.readMessage(conn)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		if msg["type"].(string) == "ping" {
			lastPingTime = time.Now()
			err := h.handlePing(conn, userGameLocation)
			if err != nil {
				log.Printf("Error handling ping: %v", err)
				break
			}
			continue
		}
		err = h.processMessage(userGameLocation, msg)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			break
		}
	}
}

func (h *UserGameLocationHandler) readMessage(conn *websocket.Conn) (map[string]interface{}, error) {

	var msg = map[string]interface{}{}
	err := conn.ReadJSON(&msg)
	if err != nil {
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

func (h *UserGameLocationHandler) processMessage(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	var err error
	switch msg["type"].(string) {
	case "join-game":
		err = h.handleJoinGame(userGameLocation, msg)
	case "join-audio":
		err = h.handleJoinAudio(userGameLocation, msg)
	case "leave-game":
		err = h.handleLeaveGame(userGameLocation, msg)
	case "leave-audio":
		err = h.handleLeaveAudio(userGameLocation, msg)
	case "move":
		err = h.handleMoveGame(userGameLocation, msg)
	case "offer", "answer", "ice-candidate":
		err = h.handleSignalingMessage(userGameLocation, msg)
	default:
		err = fmt.Errorf("unknown message type")
	}
	if err != nil {
		// 一時的なエラーの場合はリトライ
		if isTemporary(err) {
			time.Sleep(retryInterval)
			return h.processMessage(userGameLocation, msg)
		}
		log.Printf("Error processing message: %v", err)
	}
	return nil
}

func (h *UserGameLocationHandler) handleJoinGame(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {

	roomID := uint(msg["roomID"].(float64))
	if !isValidRoomId(roomID) {
		return fmt.Errorf("invalid roomID")
	}
	userGameLocation.RoomID = roomID

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}

	userGameLocation.UserID = fromUserID

	err := h.userGameLocationUsecase.ConnectUserGameLocation(userGameLocation)
	if err != nil {
		return fmt.Errorf("error connecting client to game: %v", err)
	}

	err = h.userGameLocationUsecase.SendGameJoinedEvent(userGameLocation)
	if err != nil {
		log.Printf("handleJoinGame: Error joining game: %v", err)
		return err
	}
	return nil
}
func (h *UserGameLocationHandler) handleJoinAudio(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	roomId := uint(msg["roomID"].(float64))
	if !isValidRoomId(roomId) {
		return fmt.Errorf("invalid roomID")
	}
	userGameLocation.RoomID = roomId

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	userGameLocation.UserID = fromUserID

	err := h.userGameLocationUsecase.ConnectUserGameLocation(userGameLocation)
	if err != nil {
		return fmt.Errorf("error connecting client to audio: %v", err)
	}

	err = h.userGameLocationUsecase.SendAudioJoinedEvent(userGameLocation)
	if err != nil {
		log.Printf("handleJoinAudio: Error joining audio: %v", err)
		return err
	}
	return nil
}

func (h *UserGameLocationHandler) handleLeaveGame(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	err := h.userGameLocationUsecase.LeaveInGame(userGameLocation, userGameLocation.RoomID)
	if err != nil {
		log.Printf("handleLeaveGame: Error leaving game: %v", err)
		return err
	}

	return nil
}

func (h *UserGameLocationHandler) handleLeaveAudio(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	roomId := uint(msg["roomID"].(float64))
	if !isValidRoomId(roomId) {
		return fmt.Errorf("invalid roomID")
	}
	userGameLocation.RoomID = roomId

	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	userGameLocation.UserID = fromUserID

	err := h.userGameLocationUsecase.LeaveInAudio(userGameLocation, userGameLocation.RoomID)
	if err != nil {
		log.Printf("handleLeaveAudio: Error leaving audio: %v", err)
		return err
	}

	return nil
}

func (h *UserGameLocationHandler) handleMoveGame(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	fromUserID := uint(msg["fromUserID"].(float64))
	if !isValidUserId(fromUserID) {
		return fmt.Errorf("invalid fromUserID")
	}
	roomID := uint(msg["roomID"].(float64))
	if !isValidRoomId(roomID) {
		return fmt.Errorf("invalid roomID")
	}

	userGameLocation.UserID = fromUserID
	userGameLocation.RoomID = roomID
	xAxis := int(msg["xAxis"].(float64))
	yAxis := int(msg["yAxis"].(float64))

	err := h.userGameLocationUsecase.MoveInGame(userGameLocation, xAxis, yAxis)
	if err != nil {
		log.Printf("Error updating and broadcasting user location: %v", err)
		return err
	}
	return nil
}

func (h *UserGameLocationHandler) handleSignalingMessage(userGameLocation *model.UserGameLocation, msg map[string]interface{}) error {
	toUserID := uint(msg["toUserID"].(float64))
	if !isValidUserId(toUserID) {
		return fmt.Errorf("invalid toUserID")
	}
	msgPayload := &model.Message{Payload: msg}
	// 特定のユーザーにメッセージを送信する(ここでルーム全員に送信するとブラウザ側でメモリエラーになる)
	err := h.userGameLocationUsecase.SendMessageToSpecificUser(userGameLocation, msgPayload, toUserID)
	if err != nil {
		log.Printf("handleSignalingMessage: Error sending message to specific user: %v", err)
		return err
	}
	return nil
}

func (h UserGameLocationHandler) handlePing(conn *websocket.Conn, userGameLocation *model.UserGameLocation) error {
	// ユーザーの接続状態を確認する
	err := h.userGameLocationUsecase.PingUserGameLocation(userGameLocation)
	if err != nil {
		log.Printf("handlePing: Error pinging user: %v", err)
		h.cleanUp(userGameLocation)
		return err
	}
	return nil
}

func (h UserGameLocationHandler) cleanUp(userGameLocation *model.UserGameLocation) {
	err := h.userGameLocationUsecase.DisconnectInAudio(userGameLocation, userGameLocation.RoomID)
	if err != nil {
		log.Printf("cleanUp: Error disconnecting audio: %v", err)
	}
	err = h.userGameLocationUsecase.DisconnectInGame(userGameLocation, userGameLocation.RoomID)
	if err != nil {
		log.Printf("cleanUp: Error disconnecting game: %v", err)
	}

}
