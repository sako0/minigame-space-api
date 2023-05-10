package model

import (
	"sync"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type UserLocation struct {
	gorm.Model
	UserID   uint
	User     *User
	AreaID   uint `gorm:"default:null"`
	Area     *Area
	RoomID   uint `gorm:"default:null"`
	Room     *Room
	XAxis    int
	YAxis    int
	JoinedAt int
	Conn     *websocket.Conn `gorm:"-"`
	Mutex    sync.Mutex      `gorm:"-"`
}

func NewUserLocationByConn(conn *websocket.Conn) *UserLocation {
	return &UserLocation{Conn: conn}
}
