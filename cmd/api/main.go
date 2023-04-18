package main

import (
	"log"
	"net/http"

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
	conn     *websocket.Conn
	roomId   string
	clientId string
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
			break
		}

		roomId, ok := msg["roomId"].(string)
		if !ok {
			log.Printf("roomId is missing")
			return
		}
		clientId, ok := msg["userId"].(string)
		if !ok {
			log.Printf("userId is missing")
			return
		}
		client.roomId = roomId
		client.clientId = clientId

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
				msg["clientId"] = clientId
				sendMessageToOtherClients(client, msg)
			}
		}
	}
}

func sendMessageToOtherClients(client *Client, msg map[string]interface{}) {
	room, ok := rooms[client.roomId]
	if !ok {
		log.Printf("Room not found: %s", client.roomId)
		return
	}

	for otherClient := range room.clients {
		if otherClient != client {
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

	room, ok := rooms[client.roomId]
	if !ok {
		log.Printf("Room not found: %s", client.roomId)
		return
	}

	for otherClient := range room.clients {
		if otherClient != client {
			connectedUserIds = append(connectedUserIds, otherClient.clientId)
		}
	}

	roomJoinedMsg := map[string]interface{}{
		"type":             "client-joined",
		"connectedUserIds": connectedUserIds,
		"clientId":         client.clientId,
	}

	err := client.conn.WriteJSON(roomJoinedMsg)
	if err != nil {
		log.Printf("Error sending client-joined event to client: %v", err)
		delete(room.clients, client)
	}
}
