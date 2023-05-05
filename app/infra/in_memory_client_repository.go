package infra

import (
	"sync"

	"github.com/sako0/minigame-space-api/app/domain/model"
	"github.com/sako0/minigame-space-api/app/domain/repository"
)

type InMemoryClientRepository struct {
	clients map[uint]map[string]*model.Client
	mu      sync.RWMutex
}

func NewInMemoryClientRepository() repository.ClientRepository {
	return &InMemoryClientRepository{clients: make(map[uint]map[string]*model.Client)}
}

func (r *InMemoryClientRepository) GetClient(userId string) (*model.Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, clients := range r.clients {
		if client, ok := clients[userId]; ok {
			return client, true
		}
	}
	return nil, false
}

func (r *InMemoryClientRepository) AddClient(client *model.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.RoomId]; !ok {
		r.clients[client.RoomId] = make(map[string]*model.Client)
	}
	r.clients[client.RoomId][client.UserId] = client
}

func (r *InMemoryClientRepository) RemoveClient(userId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, clients := range r.clients {
		delete(clients, userId)
	}
}

func (r *InMemoryClientRepository) GetAllClientsByRoomId(roomId uint) []*model.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*model.Client, 0, len(r.clients[roomId]))
	for _, client := range r.clients[roomId] {
		clients = append(clients, client)
	}
	return clients
}
