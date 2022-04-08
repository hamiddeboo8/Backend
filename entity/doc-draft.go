package entity

import (
	"time"

	"gorm.io/gorm"
)

type DocDraft struct {
	ID        uint64         `json:"ID" binding:"required" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time      `json:"created-at" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"updated-at" gorm:"default.CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"deleted-at" gorm:"default.CURRENT_TIMESTAMP;index"`

	DocNum     int       `json:"DocNum" gorm:"type:int"`
	DocDate    Date      `json:"DocDate" gorm:"foreignkey:DateID"`
	AtfNum     int       `json:"AtfNum" gorm:"type:int"`
	MinorNum   string    `json:"MinorNum" gorm:"type:nvarchar(10)"`
	Desc       string    `json:"Desc" gorm:"nvarchar(100)"`
	State      string    `json:"State" default:"موقت" gorm:"nvarvhar(10)"`
	DailyNum   int       `json:"DailyNum" gorm:"int"`
	DocType    string    `json:"DocType" gorm:"nvarvhar(30)"`
	EmitSystem string    `json:"EmitSystem" gorm:"nvarchar(50)"`
	DocItems   []DocItem `json:"DocItems" gorm:"foreignKey:DocRefer"`

	DateID uint64 `json:"-"`
}
