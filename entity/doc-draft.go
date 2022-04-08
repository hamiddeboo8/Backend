package entity

import (
	"time"

	"gorm.io/gorm"
)

type DocDraft struct {
	ID        uint64         `json:"ID" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time      `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"default.CURRENT_TIMESTAMP;index"`

	DocNum     int            `json:"DocNum" gorm:"type:int"`
	Year       int            `gorm:"type:int"`
	Month      int            `gorm:"type:int;"`
	Day        int            `gorm:"type:int;"`
	Hour       int            `gorm:"type:int;"`
	Minute     int            `gorm:"type:int;"`
	Second     int            `gorm:"type:int;"`
	AtfNum     int            `json:"AtfNum" gorm:"type:int"`
	MinorNum   string         `json:"MinorNum" gorm:"type:VARCHAR(10)"`
	Desc       string         `json:"Desc" gorm:"VARCHAR(100)"`
	State      string         `json:"State" default:"موقت" gorm:"VARCHAR(10)"`
	DailyNum   int            `json:"DailyNum" gorm:"int"`
	DocType    string         `json:"DocType" gorm:"VARCHAR(30)"`
	EmitSystem string         `json:"EmitSystem" gorm:"VARCHAR(50)"`
	DocItems   []DocItemDraft `json:"DocItems" gorm:"foreignKey:DocDraftRefer"`
}
