package entity

import "time"

type Tafsili struct {
	ID        uint64    `json:"ID" gorm:"primaryKey;auto_increment"`
	CreatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"-" gorm:"default.CURRENT_TIMESTAMP"`

	CodeVal       string `json:"CodeVal" gorm:"type:varchar(10);UNIQUE"`
	Name          string `json:"Name" gorm:"type:varchar(50)"`
	CurrPossible  bool   `json:"Curr" gorm:"type:boolean"`
	TrackPossible bool   `json:"Track" gorm:"type:boolean"`
}
