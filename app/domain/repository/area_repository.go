package repository

import "github.com/sako0/minigame-space-api/app/domain/model"

type AreaRepository = interface {
	GetArea(areaId uint) (*model.Area, error)
	AddArea(area *model.Area) error
	RemoveArea(areaId uint) error
}
