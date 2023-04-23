package model

import "github.com/gorilla/websocket"

type Client struct {
	conn   *websocket.Conn
	roomId string
	userId string
}

func NewClient(conn *websocket.Conn, roomId, userId string) *Client {
	return &Client{conn: conn, roomId: roomId, userId: userId}
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

func (c *Client) RoomId() string {
	return c.roomId
}

func (c *Client) UserId() string {
	return c.userId
}
