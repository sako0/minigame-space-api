package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewSQLConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db.Exec("SET time_zone = '+09:00'")
	return db, err
}
