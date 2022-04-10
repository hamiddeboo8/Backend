package entity

import "gorm.io/gorm"

type GlobalVars struct {
	gorm.Model

	TodayCount   int `gorm:"type:int"`
	AtfNumGlobal int `gorm:"type:int"`
}
