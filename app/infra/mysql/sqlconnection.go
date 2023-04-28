package infra

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewSQLConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}
