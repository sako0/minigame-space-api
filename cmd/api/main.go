package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
	wsHandler := handler.NewWebSocketHandler(*roomUsecase, upgrader)

	e := echo.New()

	e.GET("/signaling", func(c echo.Context) error {
		wsHandler.HandleConnections(c.Response().Writer, c.Request())
		return nil
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	log.Println("Starting server on :5500")
	err := e.Start(":5500")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
