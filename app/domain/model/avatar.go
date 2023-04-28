package model

import (
	"gorm.io/gorm"
)

type Avatar struct {
	gorm.Model
	Name string
	URL  string
}
