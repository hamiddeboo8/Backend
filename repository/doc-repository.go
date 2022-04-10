package repository

import (
	"AccountingDoc/Gin-Server/entity"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DocRepository interface {
	Save(entity.Doc) error
	Update(entity.Doc) (entity.Doc, error)
	Delete(entity.Doc) error
	FindAll() ([]entity.Doc, error)
	SaveDraft(entity.DocDraft) (entity.DocDraft, error)
	UpdateDraft(entity.DocDraft) (entity.DocDraft, error)
	DeleteDraft(entity.DocDraft) error
	FindAllDraft() ([]entity.DocDraft, error)
	SaveGlobal(entity.GlobalVars) (entity.GlobalVars, error)
	UpdateGlobal(entity.GlobalVars) (entity.GlobalVars, error)
	FindGlobal() (entity.GlobalVars, error)

	FindByID(id uint64) (entity.Doc, error)
	FindDraftByID(id uint64) (entity.DocDraft, error)

	DeleteByID(id uint64) error
	DeleteDraftByID(id uint64) error

	FindMoeins() ([]entity.Moein, error)
	FindTafsilis() ([]entity.Tafsili, error)

	CloseDB() error
}

type database struct {
	connection *gorm.DB
}

func NewDocRepository() DocRepository {
	dsn := "host=localhost user=postgres password=123581321345589144Hamidreza. dbname=AccountingDocs port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error in openning Database")
	}

	db.AutoMigrate(&entity.GlobalVars{}, &entity.Doc{}, &entity.DocDraft{}, &entity.DocItem{}, &entity.DocItemDraft{})

	if db.Migrator().HasTable(&entity.GlobalVars{}) {
		db.Migrator().DropTable(&entity.GlobalVars{})
	}
	if db.Migrator().HasTable(&entity.Doc{}) {
		db.Migrator().DropTable(&entity.Doc{})
	}
	if db.Migrator().HasTable(&entity.DocDraft{}) {
		db.Migrator().DropTable(&entity.DocDraft{})
	}
	if db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().DropTable(&entity.DocItem{})
	}
	if db.Migrator().HasTable(&entity.DocItemDraft{}) {
		db.Migrator().DropTable(&entity.DocItemDraft{})
	}

	if !db.Migrator().HasTable(&entity.Moein{}) {
		db.Migrator().CreateTable(&entity.Moein{})
	}
	if !db.Migrator().HasTable(&entity.Tafsili{}) {
		db.Migrator().CreateTable(&entity.Tafsili{})
	}
	if !db.Migrator().HasTable(&entity.GlobalVars{}) {
		db.Migrator().CreateTable(&entity.GlobalVars{})
		glb := &entity.GlobalVars{}
		glb.AtfNumGlobal = 1
		glb.TodayCount = 1
		db.Create(glb)
	}
	if !db.Migrator().HasTable(&entity.Doc{}) {
		db.Migrator().CreateTable(&entity.Doc{})
	}
	if !db.Migrator().HasTable(&entity.DocDraft{}) {
		db.Migrator().CreateTable(&entity.DocDraft{})
	}
	if !db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().CreateTable(&entity.DocItem{})
	}
	if !db.Migrator().HasTable(&entity.DocItemDraft{}) {
		db.Migrator().CreateTable(&entity.DocItemDraft{})
	}

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
func (db *database) Update(doc entity.Doc) (entity.Doc, error) {
	x := db.connection.Where("doc_refer = ?", doc.ID).Delete(&entity.DocItem{})
	err := x.Error
	if err != nil {
		return entity.Doc{}, err
	}
	//fmt.Println(doc)
	res := db.connection.Save(&doc)
	return doc, res.Error
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
func (db *database) SaveDraft(docDraft entity.DocDraft) (entity.DocDraft, error) {
	res := db.connection.Create(&docDraft)
	return docDraft, res.Error
}
func (db *database) UpdateDraft(docDraft entity.DocDraft) (entity.DocDraft, error) {
	x := db.connection.Where("doc_draft_refer = ?", docDraft.ID).Delete(&entity.DocItemDraft{})
	err := x.Error
	if err != nil {
		return entity.DocDraft{}, err
	}
	fmt.Println(docDraft)
	res := db.connection.Save(&docDraft)
	return docDraft, res.Error
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
func (db *database) SaveGlobal(x entity.GlobalVars) (entity.GlobalVars, error) {
	res := db.connection.Create(&x)
	return x, res.Error
}
func (db *database) UpdateGlobal(x entity.GlobalVars) (entity.GlobalVars, error) {
	res := db.connection.Save(&x)
	return x, res.Error
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
	var x []entity.DocItem
	db.connection.Where("doc_refer = ?", id).Preload("Moein").Preload("Tafsili").Find(&x)
	doc.DocItems = x
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
	var x []entity.DocItemDraft
	db.connection.Where("doc_draft_refer = ?", id).Preload("Moein").Preload("Tafsili").Find(&x)
	doc.DocItems = x
	return doc, nil
}
func (db *database) DeleteByID(id uint64) error {
	res := db.connection.Delete(&entity.Doc{}, id)
	return res.Error
}
func (db *database) DeleteDraftByID(id uint64) error {
	res := db.connection.Delete(&entity.DocDraft{}, id)
	return res.Error
}

func (db *database) FindMoeins() ([]entity.Moein, error) {
	var codes []entity.Moein
	res := db.connection.Set("gorm:auto_preload", true).Find(&codes)
	return codes, res.Error
}
func (db *database) FindTafsilis() ([]entity.Tafsili, error) {
	var codes []entity.Tafsili
	res := db.connection.Set("gorm:auto_preload", true).Find(&codes)
	return codes, res.Error
}
