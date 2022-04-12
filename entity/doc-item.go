package entity

import (
	"time"
)

type DocItem struct {
	ID        uint64    `json:"ID" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`

	Num       int     `json:"Num" binding:"required" gorm:"type:int"`
	Moein     Moein   `json:"Moein" gorm:"foreignKey:MoeinID;constraint:OnUpdate:CASCADE;"`
	MoeinID   uint64  `json:"-"`
	Tafsili   Tafsili `json:"Tafsili" gorm:"foreignKey:TafsiliID;constraint:OnUpdate:CASCADE;"`
	TafsiliID uint64  `json:"-"`
	Bedehkar  int     `json:"Bedehkar" gorm:"type:int"`
	Bestankar int     `json:"Bestankar" gorm:"type:int"`
	Desc      string  `json:"Desc" gorm:"type:VARCHAR(100)"`
	CurrPrice int     `json:"CurrPrice" gorm:"type:int"`
	Curr      string  `json:"Curr" gorm:"type:VARCHAR(20)"`
	CurrRate  int     `json:"CurrRate" gorm:"type:int"`

	SaveDB bool `json:"SaveDB" gorm:"type:boolean"`

	DocRefer uint64 `json:"-" gorm:""`
}
