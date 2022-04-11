package service

import (
	"AccountingDoc/Gin-Server/entity"
	"errors"
	"fmt"

	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocService interface {
	FindAll() ([]entity.Doc, error)
	Save(doc entity.Doc) error
	SaveByID(id uint64, doc entity.AddRemoveDocItem) error
	FindByID(id uint64) (entity.Doc, error)
	ChangeState(id uint64) error
	CanEdit(id uint64) error
	InitialCreate() (entity.Doc, error)
	ChangeIsChange(id uint64) error

	/*FindDraftByID(id uint64) (entity.DocDraft, error)
	CanEditDraft(id uint64) error
	CreateDraftDoc() (entity.DocDraft, error)
	FindDrafts() ([]entity.DocDraft, error)
	SaveDraftByID(uint64, entity.AddRemoveDocDraftItem) error
	RemoveDraft(id uint64) error*/

	FindMoeins() ([]entity.Moein, error)
	FindTafsilis() ([]entity.Tafsili, error)

	CloseDB() error
}

type docService struct {
	db *gorm.DB
}

type TransactionFunc func(tx *gorm.DB) error

func NewDbConnection() *gorm.DB {
	dsn := "host=localhost user=postgres password=123581321345589144Hamidreza. dbname=AccountingDocs port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error in openning Database")
	}

	db.AutoMigrate(&entity.GlobalVars{}, &entity.Doc{}, &entity.DocItem{} /*, &entity.DocDraft{}, &entity.DocItemDraft{}*/)

	/*if db.Migrator().HasTable(&entity.GlobalVars{}) {
		db.Migrator().DropTable(&entity.GlobalVars{})
	}
	if db.Migrator().HasTable(&entity.Doc{}) {
		db.Migrator().DropTable(&entity.Doc{})
	}
	if db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().DropTable(&entity.DocItem{})
	}*/
	/*if db.Migrator().HasTable(&entity.DocDraft{}) {
		db.Migrator().DropTable(&entity.DocDraft{})
	}
	if db.Migrator().HasTable(&entity.DocItemDraft{}) {
		db.Migrator().DropTable(&entity.DocItemDraft{})
	}*/

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
	if !db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().CreateTable(&entity.DocItem{})
	}
	/*if !db.Migrator().HasTable(&entity.DocDraft{}) {
		db.Migrator().CreateTable(&entity.DocDraft{})
	}
	if !db.Migrator().HasTable(&entity.DocItemDraft{}) {
		db.Migrator().CreateTable(&entity.DocItemDraft{})
	}*/

	return db
}

//difference of doing everything on DiInTransaction
func (service *docService) DoInTransaction(fn TransactionFunc) error {
	tx := service.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := fn(tx)
	if err != nil {
		if e := tx.Rollback().Error; e != nil {
			return e
		}
		return err
	}
	fmt.Println("Mireser")
	if err = tx.Commit().Error; err != nil {
		return err
	}
	fmt.Println("Nemirese")

	return nil
}

func (service *docService) CloseDB() error {
	postgresDB, err := service.db.DB()
	if err != nil {
		return errors.New("error in generic database")
	}
	err = postgresDB.Close()
	if err != nil {
		return errors.New("error in closing database")
	}
	return nil
}

func New(database *gorm.DB) DocService {
	return &docService{
		db: database,
	}
}

func (service *docService) FindAll() ([]entity.Doc, error) {
	var docs []entity.Doc
	res := service.db.Set("gorm:auto_preload", true).Find(&docs)
	if res.Error != nil {
		return docs, res.Error
	}
	return docs, nil
}

func (service *docService) FindByID(id uint64) (entity.Doc, error) {
	var doc entity.Doc
	res := service.db.Preload("DocItems").Preload("DocItems.Moein").Preload("DocItems.Tafsili").First(&doc, id)
	if res.Error != nil {
		return entity.Doc{}, res.Error
	}
	return doc, nil

	/*var x []entity.DocItem
	db.connection.Where("doc_refer = ?", id).Preload("Moein").Preload("Tafsili").Find(&x)
	doc.DocItems = x*/

}

func (service *docService) InitialCreate() (entity.Doc, error) {
	var doc entity.Doc
	pt := ptime.Now()
	doc.Year = pt.Year()
	doc.Month = int(pt.Month())
	doc.Day = pt.Day()
	doc.Hour = pt.Hour()
	doc.Minute = pt.Minute()
	doc.Second = pt.Second()
	doc.AtfNum = 0
	doc.DocNum = 1
	doc.DailyNum = 0
	doc.MinorNum = ""
	doc.Desc = ""
	doc.State = "موقت"
	doc.DocType = "عمومی"
	doc.EmitSystem = "سیستم حسابداری"
	doc.DocItems = make([]entity.DocItem, 0)
	return doc, nil
}

//preload???
func (service *docService) Save(doc entity.Doc) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		/*res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&docDraft, doc.ID)
		if res.Error != nil {
			return res.Error
		}
		res = tx.Delete(&entity.DocDraft{}, doc.ID)
		if res.Error != nil {
			return res.Error
		}*/
		var glb entity.GlobalVars
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&glb)
		if res.Error != nil {
			return res.Error
		}
		doc.AtfNum = glb.AtfNumGlobal
		doc.DailyNum = glb.TodayCount
		res = tx.Model(&entity.GlobalVars{}).Where("1 = 1").Updates(map[string]interface{}{"today_count": gorm.Expr("today_count + ?", 1), "atf_num_global": gorm.Expr("atf_num_global + ?", 1)})
		if res.Error != nil {
			return res.Error
		}
		res = tx.Omit("DocItems").Create(&doc)
		if res.Error != nil {
			return res.Error
		}
		for i := range doc.DocItems {
			doc.DocItems[i].DocRefer = doc.ID
			doc.DocItems[i].ID = 0
		}
		res = tx.Create(&doc.DocItems)
		return res.Error
	})
	return err
}

//preload???
func (service *docService) SaveByID(id uint64, doc entity.AddRemoveDocItem) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.Doc{}).Where("id = ?", id).Updates(map[string]interface{}{
			"doc_num":     doc.DocNum,
			"year":        doc.Year,
			"month":       doc.Month,
			"day":         doc.Day,
			"hour":        doc.Hour,
			"minute":      doc.Minute,
			"second":      doc.Second,
			"atf_num":     doc.AtfNum,
			"minor_num":   doc.MinorNum,
			"desc":        doc.Desc,
			"state":       doc.State,
			"daily_num":   doc.DailyNum,
			"doc_type":    doc.DocType,
			"emit_system": doc.EmitSystem,
		})
		if res.Error != nil {
			return res.Error
		}
		for i := range doc.AddDocItems {
			doc.AddDocItems[i].DocRefer = id
			doc.AddDocItems[i].ID = 0
		}
		if len(doc.AddDocItems) > 0 {
			res = tx.Create(&doc.AddDocItems)
			if res.Error != nil {
				return res.Error
			}
		}
		if len(doc.RemoveDocItems) > 0 {
			res = tx.Delete(&entity.DocItem{}, doc.RemoveDocItems)
			if res.Error != nil {
				return res.Error
			}
		}
		res = tx.Model(&entity.Doc{}).Where("id = ?", id).Update("is_changing", false)
		return res.Error
	})
	return err
}

func (service *docService) ChangeIsChange(id uint64) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.Doc{}).Where("id = ?", id).Update("is_changing", false)
		return res.Error
	})
	return err
}

func (service *docService) ChangeState(id uint64) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.Doc{}).Where("id = ?", id).Update("state", "دائمی")
		return res.Error
	})
	//if doc.State == "دائمی" {
	//	return doc, errors.New("doc state is permanent")
	//}
	return err
}

//select state and ischange
func (service *docService) CanEdit(id uint64) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		var doc entity.Doc
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Model(&entity.Doc{}).Where("id = ?", id).First(&doc)
		if res.Error != nil {
			fmt.Println("XXX")
			return res.Error
		}
		if doc.State == "دائمی" {
			return errors.New("doc is permanent")
		}
		fmt.Println(doc.IsChanging)
		if doc.IsChanging {
			fmt.Println("miad")
			return errors.New("sb else is changing")
		}
		res = tx.Model(&entity.Doc{}).Where("id = ?", id).Update("is_changing", true)
		fmt.Println("errr: ", res.Error)
		return res.Error
	})
	return err
}

func (service *docService) FindMoeins() ([]entity.Moein, error) {
	var codes []entity.Moein
	res := service.db.Find(&codes)
	return codes, res.Error
}
func (service *docService) FindTafsilis() ([]entity.Tafsili, error) {
	var codes []entity.Tafsili
	res := service.db.Find(&codes)
	return codes, res.Error
}

/*
//not implement yet
func (service *docService) CanEditDraft(id uint64) error {
	//multi client
	return nil
}
*/
/*
func (service *docService) SaveDraftByID(id uint64, doc entity.AddRemoveDocDraftItem) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		res := tx.Model(&entity.DocDraft{}).Where("id = ?", id).Updates(map[string]interface{}{
			"doc_num":     doc.DocNum,
			"year":        doc.Year,
			"month":       doc.Month,
			"day":         doc.Day,
			"hour":        doc.Hour,
			"minute":      doc.Minute,
			"second":      doc.Second,
			"atf_num":     doc.AtfNum,
			"minor_num":   doc.MinorNum,
			"desc":        doc.Desc,
			"state":       doc.State,
			"daily_num":   doc.DailyNum,
			"doc_type":    doc.DocType,
			"emit_system": doc.EmitSystem,
		})
		if res.Error != nil {
			return res.Error
		}
		for i, _ := range doc.AddDocItems {
			doc.AddDocItems[i].DocDraftRefer = id
		}
		res = tx.Model(&entity.DocItemDraft{}).Create(doc.AddDocItems)
		if res.Error != nil {
			return res.Error
		}
		res = tx.Delete(&entity.DocItemDraft{}, doc.RemoveDocItems)
		return res.Error
	})
	return err
}
*/

/*
//lock? no i guess
func (service *docService) RemoveDraft(id uint64) error {
	res := service.db.Delete(&entity.DocDraft{}, id)
	if res.Error != nil {
		return errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	return nil
}
*/

/*
//doubt about preload
func (service *docService) FindDrafts() ([]entity.DocDraft, error) {
	var docs []entity.DocDraft
	res := service.db.Set("gorm:auto_preload", true).Find(&docs)
	if res.Error != nil {
		return docs, res.Error
	}
	return docs, nil
}
*/
/*
func (service *docService) FindDraftByID(id uint64) (entity.DocDraft, error) {
	var doc entity.DocDraft
	res := service.db.Preload("DocItems").Preload("DocItems.Moein").Preload("DocItems.Tafsili").First(&doc, id)
	if res.Error != nil {
		return entity.DocDraft{}, res.Error
	}
	return doc, errors.New("Doc with id " + strconv.FormatUint(id, 10) + " Not Found")

	//var x []entity.DocItemDraft
	//db.connection.Where("doc_draft_refer = ?", id).Preload("Moein").Preload("Tafsili").Find(&x)
	//doc.DocItems = x
}
*/

/*
func (service *docService) CreateDraftDoc() (entity.DocDraft, error) {
	var doc entity.DocDraft
	pt := ptime.Now()
	doc.Year = pt.Year()
	doc.Month = int(pt.Month())
	doc.Day = pt.Day()
	doc.Hour = pt.Hour()
	doc.Minute = pt.Minute()
	doc.Second = pt.Second()
	doc.AtfNum = -1
	doc.DocNum = 0
	doc.DailyNum = -1
	doc.MinorNum = ""
	doc.Desc = ""
	doc.State = "موقت"
	doc.DocType = "عمومی"
	doc.EmitSystem = "سیستم حسابداری"
	doc.DocItems = make([]entity.DocItemDraft, 0)
	res := service.db.Create(&doc)
	if res.Error != nil {
		return entity.DocDraft{}, res.Error
	}
	return doc, nil
}
*/
