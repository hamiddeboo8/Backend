package repository

import (
	"AccountingDoc/Gin-Server/entity"
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DocRepository interface {
	Save(entity.Doc) error
	Update(entity.Doc) error
	Delete(entity.Doc) error
	FindAll() ([]entity.Doc, error)
	SaveDraft(entity.DocDraft) error
	UpdateDraft(entity.DocDraft) error
	DeleteDraft(entity.DocDraft) error
	FindAllDraft() ([]entity.DocDraft, error)
	SaveGlobal(entity.GlobalVars) error
	UpdateGlobal(entity.GlobalVars) error
	FindGlobal() (entity.GlobalVars, error)

	FindByID(id uint64) (entity.Doc, error)
	FindDraftByID(id uint64) (entity.DocDraft, error)

	CloseDB() error
}

type database struct {
	connection *gorm.DB
}

func NewDocRepository() DocRepository {
	dsn := "host=localhost user=postgres password=123581321345589144Hamidreza. dbname=Gorm_Train port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error in openning Database")
	}

	db.AutoMigrate(&entity.Date{}, &entity.GlobalVars{}, &entity.DocItem{}, &entity.Doc{}, &entity.DocDraft{})

	return &database{
		connection: db,
	}
}

func (db *database) CloseDB() error {
	postgresDB, err := db.connection.DB()
	if err != nil {
		return errors.New("error in generic database")
	}
	err = postgresDB.Close()
	if err != nil {
		return errors.New("error in closing database")
	}
	return nil
}

func (db *database) Save(doc entity.Doc) error {
	res := db.connection.Create(&doc)
	return res.Error
}
func (db *database) Update(doc entity.Doc) error {
	res := db.connection.Save(&doc)
	return res.Error
}
func (db *database) Delete(doc entity.Doc) error {
	res := db.connection.Delete(&doc)
	return res.Error
}
func (db *database) FindAll() ([]entity.Doc, error) {
	var docs []entity.Doc
	res := db.connection.Set("gorm:auto_preload", true).Find(&docs)
	return docs, res.Error
}
func (db *database) SaveDraft(docDraft entity.DocDraft) error {
	res := db.connection.Create(&docDraft)
	return res.Error
}
func (db *database) UpdateDraft(docDraft entity.DocDraft) error {
	res := db.connection.Save(&docDraft)
	return res.Error
}
func (db *database) DeleteDraft(docDraft entity.DocDraft) error {
	res := db.connection.Delete(&docDraft)
	return res.Error
}
func (db *database) FindAllDraft() ([]entity.DocDraft, error) {
	var docDrafts []entity.DocDraft
	res := db.connection.Set("gorm:auto_preload", true).Find(&docDrafts)
	return docDrafts, res.Error
}
func (db *database) SaveGlobal(x entity.GlobalVars) error {
	res := db.connection.Create(&x)
	return res.Error
}
func (db *database) UpdateGlobal(x entity.GlobalVars) error {
	res := db.connection.Save(&x)
	return res.Error
}
func (db *database) FindGlobal() (entity.GlobalVars, error) {
	var x entity.GlobalVars
	res := db.connection.Set("gorm:auto_preload", true).Find(&x)
	return x, res.Error
}

func (db *database) FindByID(id uint64) (entity.Doc, error) {
	var doc entity.Doc
	res := db.connection.First(&doc, id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return entity.Doc{}, errors.New("doc not found")
		}
		return entity.Doc{}, res.Error
	}
	return doc, nil

}
func (db *database) FindDraftByID(id uint64) (entity.DocDraft, error) {
	var doc entity.DocDraft
	res := db.connection.First(&doc, id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return entity.DocDraft{}, errors.New("docDraft not found")
		}
		return entity.DocDraft{}, res.Error
	}
	return doc, nil
}
