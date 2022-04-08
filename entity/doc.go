package entity

import (
	"time"

	"gorm.io/gorm"
)

type Doc struct {
	ID        uint64         `json:"ID" binding:"required" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time      `json:"created-at" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"updated-at" gorm:"default.CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"deleted-at" gorm:"default.CURRENT_TIMESTAMP;index"`

	DocNum     int       `json:"DocNum" binding:"required" gorm:"type:int"`
	DocDate    Date      `json:"DocDate" binding:"required" gorm:"foreignkey:DateID"`
	AtfNum     int       `json:"AtfNum" binding:"required" gorm:"type:int"`
	MinorNum   string    `json:"MinorNum" gorm:"type:nvarchar(10)"`
	Desc       string    `json:"Desc" gorm:"nvarchar(100)"`
	State      string    `json:"State" default:"موقت" gorm:"nvarvhar(10)"`
	DailyNum   int       `json:"DailyNum" gorm:"int"`
	DocType    string    `json:"DocType" binding:"required" gorm:"nvarvhar(30)"`
	EmitSystem string    `json:"EmitSystem" binding:"required" gorm:"nvarchar(50)"`
	DocItems   []DocItem `json:"DocItems" binding:"required" gorm:"foreignKey:DocRefer"`

	DateID uint64 `json:"-"`
}
