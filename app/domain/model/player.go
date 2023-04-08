package model

import (
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	UserID string `json:"userID"`
	Conn   *websocket.Conn
}
