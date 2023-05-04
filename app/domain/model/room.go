package model

type Room struct {
	ID      uint
	Clients []Client
}

func NewRoom(id uint) *Room {
	return &Room{ID: id}
}
