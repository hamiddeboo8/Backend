package entity

type Date struct {
	Year   int `gorm:"type:int"`
	Month  int `gorm:"type:int;check:Month<=12,Month>=1"`
	Day    int `gorm:"type:int;check:Day<=31,Day>=1"`
	Hour   int `gorm:"type:int;check:Hour<=24,Hour>=0"`
	Minute int `gorm:"type:int;check:Minute<=60,Minute>=0"`
	Second int `gorm:"type:int;check:Second<=60,Second>=0"`
}
