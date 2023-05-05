package model

import "github.com/gorilla/websocket"

type Client struct {
	Conn   *websocket.Conn
	RoomId uint
	Room   *Room
	UserId string
}

func NewClient(conn *websocket.Conn, room *Room, userId string) *Client {
	return &Client{Conn: conn, Room: room, UserId: userId}
}

func NewClientByConn(conn *websocket.Conn) *Client {
	return &Client{Conn: conn}
}
