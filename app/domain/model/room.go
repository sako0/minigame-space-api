package model

type Room struct {
	Clients map[*Client]bool
}

func NewRoom() *Room {
	return &Room{Clients: make(map[*Client]bool)}
}
