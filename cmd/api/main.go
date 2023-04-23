package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sako0/minigame-space-api/app/infra"
	"github.com/sako0/minigame-space-api/app/usecase"
	handler "github.com/sako0/minigame-space-api/app/websocket"
)

func main() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	roomRepo := infra.NewInMemoryRoomRepository()
	roomUsecase := usecase.NewRoomUsecase(roomRepo)
	handler := handler.NewWebSocketHandler(*roomUsecase, upgrader)

	http.HandleFunc("/socket.io/", handler.HandleConnections)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Println("Starting server on :5500")
	err := http.ListenAndServe(":5500", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
