package main

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Room struct {
	ID      string
	Players []*websocket.Conn
}

var rooms = make(map[string]*Room)
var roomsLock = &sync.RWMutex{}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/ws/:roomID", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		roomID := c.Param("roomID")
		roomsLock.Lock()
		room, ok := rooms[roomID]
		if !ok {
			room = &Room{ID: roomID}
			rooms[roomID] = room
		}
		room.Players = append(room.Players, conn)
		roomsLock.Unlock()

		for {
			msg := &Message{}
			err := conn.ReadJSON(msg)
			if err != nil {
				break
			}

			// ゲームロジックを実装
			switch msg.Type {
			case "chat":
				// チャットメッセージをブロードキャスト
				for _, player := range room.Players {
					if player != conn {
						player.WriteJSON(msg)
					}
				}
				// その他のゲームロジックを実装
			}
		}

		// プレイヤーをルームから削除
		roomsLock.Lock()
		for i, player := range room.Players {
			if player == conn {
				room.Players = append(room.Players[:i], room.Players[i+1:]...)
				break
			}
		}
		roomsLock.Unlock()

		return nil
	})

	e.Logger.Fatal(e.Start(":5500"))
}
