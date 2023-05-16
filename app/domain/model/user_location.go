package model

import (
	"encoding/json"
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

func (u *UserLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserID uint `json:"userID"`
		AreaID uint `json:"areaID"`
		RoomID uint `json:"roomID"`
		XAxis  int  `json:"xAxis"`
		YAxis  int  `json:"yAxis"`
	}{
		UserID: u.UserID,
		AreaID: u.AreaID,
		RoomID: u.RoomID,
		XAxis:  u.XAxis,
		YAxis:  u.YAxis,
	})
}
