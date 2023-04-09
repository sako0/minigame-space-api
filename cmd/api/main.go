package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Create an Echo server instance
	e := echo.New()

	// Map to store the connected clients in each room
	roomClients := make(map[string]map[*websocket.Conn]bool)

	// Create a WebSocket handler function
	websocketHandler := func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		defer func() {
			for roomId, clients := range roomClients {
				if _, ok := clients[ws]; ok {
					delete(clients, ws)
					if len(clients) == 0 {
						delete(roomClients, roomId)
					}
				}
			}
			ws.Close()
		}()

		// Log the connection
		log.Println("connected:", ws.RemoteAddr())

		// Read incoming messages
		for {
			var message struct {
				Type      string      `json:"type"`
				RoomId    string      `json:"roomId"`
				Data      interface{} `json:"data"`
				Candidate interface{} `json:"candidate"`
				SDP       interface{} `json:"sdp"`
			}
			if err := ws.ReadJSON(&message); err != nil {
				log.Println("error receiving message:", err)
				return
			}

			if message.Type == "join" {
				// Add the client to the list of clients in the room
				roomId := message.RoomId
				if roomId == "" {
					log.Println("Invalid room ID:", roomId)
					return
				}
				if _, ok := roomClients[roomId]; !ok {
					roomClients[roomId] = make(map[*websocket.Conn]bool)
				}
				roomClients[roomId][ws] = true

				// Log the room join
				log.Println("client joined room:", roomId)
			} else if message.Type == "signal" {
				// Broadcast the signal message to other clients in the same room
				roomId := message.RoomId
				if clients, ok := roomClients[roomId]; ok {
					for client := range clients {
						if client == ws {
							continue
						}
						if err := client.WriteJSON(message); err != nil {
							log.Println("error sending message:", err)
						}
					}
				}

				// Log the signal message
				log.Println("signal message sent to room:", roomId)
			} else if message.Type == "candidate" {
				// Broadcast the candidate message to other clients in the same room
				roomId := message.RoomId
				if clients, ok := roomClients[roomId]; ok {
					for client := range clients {
						if client == ws {
							continue
						}
						if err := client.WriteJSON(message); err != nil {
							log.Println("error sending message:", err)
						}
					}
				}

				// Log the candidate message
				log.Println("candidate message sent to room:", roomId)
			} else if message.Type == "offer" {
				// Broadcast the offer message to other clients in the same room
				roomId := message.RoomId
				if clients, ok := roomClients[roomId]; ok {
					for client := range clients {
						if client == ws {
							continue
						}
						if err := client.WriteJSON(message); err != nil {
							log.Println("error sending message:", err)
						}
					}
				}

				// Log the offer message
				log.Println("offer message sent to room:", roomId)

			} else if message.Type == "answer" {
				// Broadcast the answer message to other clients in the same room
				roomId := message.RoomId
				if clients, ok := roomClients[roomId]; ok {
					for client := range clients {
						if client == ws {
							continue
						}
						if err := client.WriteJSON(message); err != nil {
							log.Println("error sending message:", err)
						}
					}
				}

				// Log the answer message
				log.Println("answer message sent to room:", roomId)
			}
		}
	}

	// Register the WebSocket handler
	e.GET("/socket.io/", func(c echo.Context) error {
		websocketHandler(c.Response().Writer, c.Request())
		return nil
	})

	// Serve the Echo server
	e.Logger.Fatal(e.Start(":5500"))
}
