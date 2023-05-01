package repository

import (
	"github.com/sako0/minigame-space-api/app/domain/model"
)

type AreaRepository interface {
	GetArea(areaId string) (*model.Area, error)
}
