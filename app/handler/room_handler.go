package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/usecase"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type RoomHandler struct {
	RoomUsecase usecase.RoomUsecase
}

func NewRoomHandler(roomUsecase usecase.RoomUsecase) *RoomHandler {
	return &RoomHandler{
		RoomUsecase: roomUsecase,
	}
}

func (h *RoomHandler) ListRooms(c echo.Context) error {
	roomsInfo, err := h.RoomUsecase.ListRooms()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, roomsInfo)
}

func (h *RoomHandler) JoinRoom(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "failed to upgrade connection")
	}
	defer conn.Close()

	userID := c.Get("userID").(string)

	room, err := h.RoomUsecase.JoinRoom(000, userID, conn)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	for {
		msg := &model.WsMessage{}
		// uintに変換する
		targetPlayerID, err := strconv.ParseUint(msg.TargetPlayerID, 10, 64)
		if err != nil {
			break
		}
		err = conn.ReadJSON(msg)
		if err != nil {
			break
		}

		room, err := h.RoomUsecase.GetRoom(room.ID)
		if err != nil {
			break
		}

		switch msg.Type {
		case "signaling":
			h.RoomUsecase.Broadcast(room, msg)
		case "chat":
			h.RoomUsecase.Broadcast(room, msg)
		case "removePlayer":
			h.RoomUsecase.RemovePlayer(room.ID, uint(targetPlayerID))
		}
	}

	h.RoomUsecase.LeaveRoom(000, 000)

	return nil
}

func SetRoomHandler(e *echo.Echo, roomUsecase *usecase.RoomUsecase) {
	h := NewRoomHandler(*roomUsecase)

	e.GET("/rooms", h.ListRooms)
	e.GET("/rooms/:roomID", h.JoinRoom)
}
