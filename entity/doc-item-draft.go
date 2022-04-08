package entity

import (
	"time"

	"gorm.io/gorm"
)

type DocItemDraft struct {
	ID        uint64         `json:"ID" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time      `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"default.CURRENT_TIMESTAMP;index"`

	Num       int     `json:"Num" binding:"required" gorm:"type:int"`
	Moein     string  `json:"Moein" binding:"required" gorm:"type:VARCHAR(20)"`
	Tafsili   string  `json:"Tafsili" gorm:"type:VARCHAR(20)"`
	Bedehkar  int     `json:"Bedehkar" binding:"required" gorm:"type:int"`
	Bestankar int     `json:"Bestankar" binding:"required" gorm:"type:int"`
	Desc      string  `json:"Desc" binding:"required" gorm:"type:VARCHAR(100)"`
	CurrPrice float32 `json:"CurrPrice" gorm:"type:float(8)"`
	Curr      string  `json:"Curr" gorm:"type:VARCHAR(20)"`
	CurrRate  float32 `json:"CurrRate" gorm:"type:float(8)"`

	SaveDB bool `json:"SaveDB" gorm:"type:boolean"`

	DocDraftRefer uint64 `json:"-"`
}
