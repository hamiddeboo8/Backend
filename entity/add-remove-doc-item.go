package entity

type AddRemoveDocItem struct {
	ID uint64 `json:"ID"`

	DocNum         int `json:"DocNum"`
	Year           int
	Month          int
	Day            int
	Hour           int
	Minute         int
	Second         int
	AtfNum         int       `json:"AtfNum"`
	MinorNum       string    `json:"MinorNum"`
	Desc           string    `json:"Desc"`
	State          string    `json:"State" default:"موقت"`
	DailyNum       int       `json:"DailyNum"`
	DocType        string    `json:"DocType" binding:"required"`
	EmitSystem     string    `json:"EmitSystem" binding:"required"`
	AddDocItems    []DocItem `json:"AddDocItems"`
	RemoveDocItems []uint64  `json:"RemoveDocItems"`
	EditDocItems   []DocItem `json:"EditDocItems"`

	IsChanging bool `json:"-" gorm:"type:boolean"`
}
