package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sako0/minigame-space-api/app/config"
	infra "github.com/sako0/minigame-space-api/app/infra/mysql"
	apiHandler "github.com/sako0/minigame-space-api/app/rest"
	"github.com/sako0/minigame-space-api/app/usecase"
	wsHandler "github.com/sako0/minigame-space-api/app/websocket"
)

func main() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// 設定読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	// データベース接続
	db, err := infra.NewSQLConnection(cfg.AppInfo.DatabaseURL)
	if err != nil {
		panic(err)
	}
	roomRepo, err := infra.NewRoomRepository(db)
	if err != nil {
		panic(err)
	}
	areaRepo, err := infra.NewAreaRepository(db)
	if err != nil {
		panic(err)
	}
	roomTypeRepo, err := infra.NewRoomTypeRepository(db)
	if err != nil {
		panic(err)
	}
	userRepo, err := infra.NewUserRepository(db)
	if err != nil {
		panic(err)
	}
	userLocationRepo, err := infra.NewUserLocationRepository(db)
	if err != nil {
		panic(err)
	}
	roomUsecase := usecase.NewRoomUsecase(roomRepo, areaRepo, roomTypeRepo, userRepo, userLocationRepo)
	wsHandler := wsHandler.NewWebSocketHandler(*roomUsecase, upgrader)
	roomApiHandler := apiHandler.NewRoomHandler(*roomUsecase)
	e := echo.New()

	e.GET("/signaling", func(c echo.Context) error {
		wsHandler.HandleConnections(c.Response().Writer, c.Request())
		return nil
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/api/rooms/create", roomApiHandler.CreateRoom)
	e.POST("/api/rooms/join", roomApiHandler.JoinRoom)

	log.Println("Starting server on :5500")
	err = e.Start(":5500")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
