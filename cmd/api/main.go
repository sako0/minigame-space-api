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
				sendMessageToOtherClients(ws, roomId, msg)
			case "answer":
				sendMessageToOtherClients(ws, roomId, msg)
				log.Printf("answerを送ったよ")
			case "candidate":
				log.Printf("candidateを送ったよ")
				sendMessageToOtherClients(ws, roomId, msg)
			}
		}
	}

}
func sendMessageToOtherClients(ws *websocket.Conn, roomID string, msg map[string]interface{}) {
	for client := range rooms[roomID].clients {
		if client != ws {
			client.WriteJSON(msg)
		}
	}
}
