package entity

import "gorm.io/gorm"

type GlobalVars struct {
	gorm.Model

	TodayCount   int `gorm:"type:int"`
	AtfNumGlobal int `gorm:"type:int"`
}

func InitGlobalVars() *GlobalVars {
	x := new(GlobalVars)
	x.AtfNumGlobal = 1
	x.TodayCount = 1
	return x
}
