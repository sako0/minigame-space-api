package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn   *websocket.Conn
	roomId string
	userId string
}

type Room struct {
	clients map[*Client]bool
}

func main() {
	http.HandleFunc("/socket.io/", handleConnections)
	// helth check
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Println("Starting server on :5500")
	err := http.ListenAndServe(":5500", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var rooms = make(map[string]*Room)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := &Client{conn: conn}

	for {
		var msg = map[string]interface{}{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			room, ok := rooms[client.roomId]
			if ok {
				delete(room.clients, client)
				if len(room.clients) == 0 {
					delete(rooms, client.roomId)
				}
			}
			break
		}

		roomId, ok := msg["roomId"].(string)
		if !ok {
			log.Printf("roomId is missing")
			return
		}
		userId, ok := msg["userId"].(string)
		if !ok {
			log.Printf("userId is missing")
			return
		}
		client.roomId = roomId
		client.userId = userId

		if _, ok := rooms[roomId]; !ok {
			rooms[roomId] = &Room{clients: make(map[*Client]bool)}
		}
		rooms[roomId].clients[client] = true

		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join-room":
				sendRoomJoinedEvent(client)
			case "offer", "answer", "ice-candidate":
				toUserId, ok := msg["toUserId"].(string)
				if !ok {
					log.Printf("toUserId is missing")
					return
				}
				msg["userId"] = userId
				sendMessageToOtherClients(client, toUserId, msg)
			}
		}
	}
}
func sendMessageToOtherClients(client *Client, toUserId string, msg map[string]interface{}) {
	room, ok := rooms[client.roomId]
	if !ok {
		log.Printf("Room not found: %s", client.roomId)
		return
	}

	for otherClient := range room.clients {
		if otherClient.userId == toUserId {
			err := otherClient.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				delete(room.clients, otherClient)
			}
		}
	}
}

func sendRoomJoinedEvent(client *Client) {
	connectedUserIds := []string{}
	time.Sleep(1 * time.Second)
	room, ok := rooms[client.roomId]
	if !ok {
		log.Printf("Room not found: %s", client.roomId)
		return
	}

	for otherClient := range room.clients {
		if otherClient != client {
			connectedUserIds = append(connectedUserIds, otherClient.userId)
		}
	}
	log.Printf("Connected user IDs: %v", connectedUserIds)
	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"userId":           client.userId,
	}

	err := client.conn.WriteJSON(roomJoinedMsg)
	if err != nil {
		log.Printf("Error sending client-joined event to client: %v", err)
		delete(room.clients, client)
	}
}
