package entity

import (
	"time"
)

type Doc struct {
	ID        uint64    `json:"ID" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`

	DocNum     int       `json:"DocNum" gorm:"type:int;UNIQUE"`
	Year       int       `gorm:"type:int"`
	Month      int       `gorm:"type:int;check:Month<=12 and Month>=1"`
	Day        int       `gorm:"type:int;check:Day<=31 and Day>=1"`
	Hour       int       `gorm:"type:int;check:Hour<=24 and Hour>=0"`
	Minute     int       `gorm:"type:int;check:Minute<=60 and Minute>=0"`
	Second     int       `gorm:"type:int;check:Second<=60 and Second>=0"`
	AtfNum     int       `json:"AtfNum" gorm:"type:int"`
	MinorNum   string    `json:"MinorNum" gorm:"type:VARCHAR(10)"`
	Desc       string    `json:"Desc" gorm:"VARCHAR(100)"`
	State      string    `json:"State" default:"موقت" gorm:"VARCHAR(10)"`
	DailyNum   int       `json:"DailyNum" gorm:"int"`
	DocType    string    `json:"DocType" binding:"required" gorm:"VARCHAR(30)"`
	EmitSystem string    `json:"EmitSystem" binding:"required" gorm:"VARCHAR(50)"`
	DocItems   []DocItem `json:"DocItems" gorm:"foreignKey:DocRefer;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	IsChanging bool `json:"-" gorm:"type:boolean"`
}
