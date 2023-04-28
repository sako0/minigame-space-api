package main

import (
	"log"

	"github.com/sako0/minigame-space-api/app/config"
	"github.com/sako0/minigame-space-api/app/domain/model"
	infra "github.com/sako0/minigame-space-api/app/infra/mysql"
)

func main() {
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
	// テーブル削除
	err = db.Migrator().DropTable(
		&model.Room{},
		&model.User{},
		&model.UserLocation{},
		&model.Avatar{},
		&model.RoomType{},
		&model.Area{},
	)
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}
	log.Println("Tables dropped")

	// テーブル作成
	err = db.AutoMigrate(
		&model.Room{},
		&model.User{},
		&model.UserLocation{},
		&model.Avatar{},
		&model.RoomType{},
		&model.Area{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
	log.Println("Migration succeeded")
}
