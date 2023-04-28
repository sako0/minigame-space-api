package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sako0/minigame-space-api/app/usecase"
)

type RoomHandler struct {
	roomUsecase usecase.RoomUsecase
}

func NewRoomHandler(roomUsecase usecase.RoomUsecase) *RoomHandler {
	return &RoomHandler{roomUsecase: roomUsecase}
}

func (h *RoomHandler) CreateRoom(c echo.Context) error {
	var req struct {
		AreaID     uint `json:"areaId"`
		RoomTypeID uint `json:"roomTypeId"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	room, err := h.roomUsecase.CreateRoom(req.AreaID, req.RoomTypeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, room)
}

func (h *RoomHandler) JoinRoom(c echo.Context) error {
	var req struct {
		FirebaseUID string `json:"firebase_uid"`
		RoomID      uint   `json:"roomId"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err := h.roomUsecase.JoinRoom(req.FirebaseUID, req.RoomID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
