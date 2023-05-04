package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type ClientRepository interface {
	GetClient(userId string) (*model.Client, bool)
	AddClient(client *model.Client)
	RemoveClient(userId string)
	GetAllClientsByRoomId(roomId uint) []*model.Client
}
