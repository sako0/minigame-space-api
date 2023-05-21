package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sako0/minigame-space-api/app/config"
	"github.com/sako0/minigame-space-api/app/database"
	"github.com/sako0/minigame-space-api/app/infra/gorm"
	"github.com/sako0/minigame-space-api/app/infra/in_memory"

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
	// 設定読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	// データベース接続
	db, err := database.NewSQLConnection(cfg.AppInfo.DatabaseURL)
	if err != nil {
		panic(err)
	}

	userLocationRepo := gorm.NewUserLocationRepository(db)
	userGameLocation := gorm.NewUserGameLocationRepository(db)
	inMemoryUserLocationRepo := in_memory.NewInMemoryUserLocationRepository()
	inMemoryUserGameLocationRepo := in_memory.NewInMemoryUserGameLocationRepository()
	roomUsecase := usecase.NewUserLocationUsecase(userLocationRepo, inMemoryUserLocationRepo)
	userGameLocationUsecase := usecase.NewUserGameLocationUsecase(userGameLocation, inMemoryUserGameLocationRepo)
	wsHandler := handler.NewWebSocketHandler(*roomUsecase, upgrader)
	wsGameHandler := handler.NewUserGameLocationHandler(*userGameLocationUsecase, upgrader)

	e := echo.New()

	e.GET("/ws", func(c echo.Context) error {
		wsHandler.HandleConnections(c.Response().Writer, c.Request())
		return nil
	})
	e.GET("/game", func(c echo.Context) error {
		wsGameHandler.HandleConnections(c.Response().Writer, c.Request())
		return nil
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	log.Println("Starting server on :5500")
	err = e.Start(":5500")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
