package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Room struct {
	clients map[*websocket.Conn]bool
}

func main() {
	http.HandleFunc("/socket.io/", handleConnections)
	log.Println("Starting server on :5500")
	err := http.ListenAndServe(":5500", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var rooms = make(map[string]*Room)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	for {
		var msg = map[string]interface{}{}

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}
		roomId, ok := msg["roomId"].(string)
		if !ok {
			log.Printf("roomIdがないよ")
			return
		}
		if _, ok := rooms[roomId]; !ok {
			rooms[roomId] = &Room{clients: make(map[*websocket.Conn]bool)}
		}
		rooms[roomId].clients[ws] = true

		log.Printf("Received message: %v", msg)
		if val, ok := msg["type"]; ok {
			switch val.(string) {
			case "join":
				log.Printf("参加したよ")
				// roomに参加する
				rooms[roomId].clients[ws] = true
			case "offer":
				log.Printf("offerを送ったよ")
				newClient := &client{conn: ws, mu: sync.Mutex{}, roomID: roomId}
				sendMessageToOtherClients(newClient, msg)
			case "answer":
				newClient := &client{conn: ws, mu: sync.Mutex{}, roomID: roomId}
				sendMessageToOtherClients(newClient, msg)
				log.Printf("answerを送ったよ")
			case "candidate":
				log.Printf("candidateを送ったよ")
				newClient := &client{conn: ws, mu: sync.Mutex{}, roomID: roomId}

				sendMessageToOtherClients(newClient, msg)
			}
		}
	}

}

type client struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	roomID string
}

var clients = make(map[*client]struct{})

func sendMessageToOtherClients(senderClient *client, message interface{}) {
	for client := range clients {
		if client != senderClient {
			client.mu.Lock()
			err := client.conn.WriteJSON(message)
			client.mu.Unlock()
			if err != nil {
				log.Printf("Error sending message: %v", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}
