package service

import (
	"AccountingDoc/Gin-Server/entity"
	"errors"
	"strconv"

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
	//InitialCreate() (entity.Doc, error)
	ChangeIsChange(id uint64) error
	DeleteByID(id uint64) error
	ValidateDocItem(entity.DocItem) error
	FilterByMinorNum(string) ([]entity.Doc, error)
	Numbering() error

	/*FindDraftByID(id uint64) (entity.DocDraft, error)
	CanEditDraft(id uint64) error
	CreateDraftDoc() (entity.DocDraft, error)
	FindDrafts() ([]entity.DocDraft, error)
	SaveDraftByID(uint64, entity.AddRemoveDocDraftItem) error
	RemoveDraft(id uint64) error*/

	FindMoeins() ([]entity.Moein, error)
	FindTafsilis() ([]entity.Tafsili, error)

	CloseDB() error

	GetDB() *gorm.DB
}

type docService struct {
	db *gorm.DB
}

type TransactionFunc func(tx *gorm.DB) error

type justID struct {
	ID uint64
}
type Date struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

func NewDbConnection() *gorm.DB {
	dsn := "host=localhost user=postgres password=123581321345589144Hamidreza. dbname=AccountingDocs port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error in openning Database")
	}

	db.AutoMigrate(&entity.GlobalVars{}, &entity.Doc{}, &entity.DocItem{})

	/*if db.Migrator().HasTable(&entity.GlobalVars{}) {
		db.Migrator().DropTable(&entity.GlobalVars{})
	}
	if db.Migrator().HasTable(&entity.Doc{}) {
		db.Migrator().DropTable(&entity.Doc{})
	}
	if db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().DropTable(&entity.DocItem{})
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

	return db
}

func (service *docService) GetDB() *gorm.DB {
	return service.db
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
	if err = tx.Commit().Error; err != nil {
		return err
	}

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
	res := service.db.Model(&entity.Doc{}).Omit("DocItems").Order("id asc").Find(&docs)
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
}

//preload???
func (service *docService) Save(doc entity.Doc) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
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
		for i := 0; i < len(doc.MinorNum); i++ {
			_, err := strconv.Atoi(string(doc.MinorNum[i]))
			if err != nil {
				return errors.New("minor num must be all digit")
			}
		}
		var date Date
		res = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.Doc{}).Where("state = ?", "??????????").Order("year desc, month desc, day desc, hour desc, minute desc, second desc").Limit(1).Find(&date)
		if res.Error != nil {
			return res.Error
		}
		dateDoc := Date{doc.Year, doc.Month, doc.Day, doc.Hour, doc.Minute, doc.Second}
		if compare(&dateDoc, &date) <= 0 {
			return errors.New("date must be bigger than latest permanent doc")
		}
		res = tx.Omit("DocItems").Create(&doc)
		if res.Error != nil {
			return res.Error
		}
		if len(doc.DocItems) == 0 {
			return errors.New("must have at least one doc item")
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

func compare(d1 *Date, d2 *Date) int {
	if d1.Year > d2.Year {
		return 1
	} else if d1.Year < d2.Year {
		return -1
	}
	if d1.Month > d2.Month {
		return 1
	} else if d1.Month < d2.Month {
		return -1
	}
	if d1.Day > d2.Day {
		return 1
	} else if d1.Day < d2.Day {
		return -1
	}
	if d1.Hour > d2.Hour {
		return 1
	} else if d1.Hour < d2.Hour {
		return -1
	}
	if d1.Minute > d2.Minute {
		return 1
	} else if d1.Minute < d2.Minute {
		return -1
	}
	if d1.Second > d2.Second {
		return 1
	} else if d1.Second < d2.Second {
		return -1
	}
	return 0

}

//preload???
func (service *docService) SaveByID(id uint64, doc entity.AddRemoveDocItem) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		var date Date
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Model(&entity.Doc{}).Where("state = ?", "??????????").Order("year desc, month desc, day desc, hour desc, minute desc, second desc").Limit(1).Find(&date)
		if res.Error != nil {
			return res.Error
		}
		dateDoc := Date{doc.Year, doc.Month, doc.Day, doc.Hour, doc.Minute, doc.Second}
		if compare(&dateDoc, &date) <= 0 {
			return errors.New("date must be bigger than latest permanent doc")
		}
		res = tx.Model(&entity.Doc{}).Where("id = ?", id).Updates(map[string]interface{}{
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

		var doc_temp entity.Doc
		tx.Model(&entity.Doc{}).Preload("DocItems").Where("id = ?", id).Find(&doc_temp)
		x := len(doc_temp.DocItems)
		if len(doc.AddDocItems) == 0 && len(doc.EditDocItems) == 0 && len(doc.RemoveDocItems) == x {
			return errors.New("it must have at least 1 doc item")
		}
		for i := range doc.AddDocItems {
			doc.AddDocItems[i].DocRefer = id
			doc.AddDocItems[i].ID = 0
		}
		for i := range doc.EditDocItems {
			doc.EditDocItems[i].DocRefer = id
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
		if len(doc.EditDocItems) > 0 {
			res = tx.Save(&doc.EditDocItems)
			if res.Error != nil {
				return res.Error
			}
		}
		return nil
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
		var doc entity.Doc
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Model(&entity.Doc{}).Where("id = ?", id).Select("state", "is_changing").First(&doc)
		if res.Error != nil {
			return res.Error
		}
		if doc.State == "??????????" {
			return errors.New("doc is permanent")
		}
		if doc.IsChanging {
			return errors.New("sb else is changing doc")
		}
		res = tx.Model(&entity.Doc{}).Where("id = ?", id).Update("state", "??????????")
		return res.Error
	})

	return err
}

//select state and ischange
func (service *docService) CanEdit(id uint64) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		var doc entity.Doc
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Model(&entity.Doc{}).Where("id = ?", id).Select("state", "is_changing").First(&doc)
		if res.Error != nil {
			return res.Error
		}
		if doc.State == "??????????" {
			return errors.New("doc is permanent")
		}
		if doc.IsChanging {
			return errors.New("sb else is changing")
		}
		res = tx.Model(&entity.Doc{}).Where("id = ?", id).Update("is_changing", true)
		return res.Error
	})
	return err
}

func (service *docService) DeleteByID(id uint64) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		var doc entity.Doc
		res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.Doc{}).Where("id = ?", id).Select("state", "is_changing").First(&doc)
		if res.Error != nil {
			return res.Error
		}
		if doc.State == "??????????" {
			return errors.New("doc is permanent")
		}
		if doc.IsChanging {
			return errors.New("sb else is changing")
		}
		var temp justID
		res = tx.Model(&entity.Doc{}).Select("id").Last(&temp)
		if res.Error != nil {
			return res.Error
		}
		if temp.ID == id {
			res = tx.Model(&entity.GlobalVars{}).Where("1 = 1").Updates(map[string]interface{}{"today_count": gorm.Expr("today_count - ?", 1), "atf_num_global": gorm.Expr("atf_num_global - ?", 1)})
			if res.Error != nil {
				return res.Error
			}
		}
		res = tx.Delete(&entity.Doc{}, id)
		return res.Error
	})
	return err
}

func (service *docService) FindMoeins() ([]entity.Moein, error) {
	var codes []entity.Moein
	res := service.db.Order("id asc").Find(&codes)
	return codes, res.Error
}
func (service *docService) FindTafsilis() ([]entity.Tafsili, error) {
	var codes []entity.Tafsili
	res := service.db.Order("id asc").Find(&codes)
	return codes, res.Error
}
func (service *docService) ValidateDocItem(docItem entity.DocItem) error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {

		if docItem.Desc == "" {
			return errors.New("docItem needs a description")
		}

		if docItem.Bedehkar == 0 && docItem.Bestankar == 0 {
			return errors.New("docItem must have value in bestankar or bedehkar")
		}
		if docItem.Bedehkar != 0 && docItem.Bestankar != 0 {
			return errors.New("docItem must have value in one of bestankar and bedehkar fields")
		}

		var moein entity.Moein
		var tafsili entity.Tafsili
		res := tx.Model(&entity.Moein{}).Where("code_val = ?", docItem.Moein.CodeVal).First(&moein)
		if res.Error != nil {
			return res.Error
		}
		if moein.TrackPossible {
			if docItem.Tafsili.CodeVal == "" {
				return errors.New("moein must have tafsili")
			}
			res = tx.Model(&entity.Tafsili{}).Where("code_val = ?", docItem.Tafsili.CodeVal).First(&tafsili)
			if res.Error != nil {
				return res.Error
			}
		} else {
			if docItem.Tafsili.CodeVal != "" {
				return errors.New("moein cannot have any tafsilis")
			}
			res = tx.Model(&entity.Tafsili{}).Where("code_val = ?", docItem.Tafsili.CodeVal).First(&tafsili)
			if res.Error != nil {
				return res.Error
			}

		}
		if moein.CurrPossible {
			if docItem.CurrPrice == 0 || docItem.Curr == "" || docItem.CurrRate == 0 {
				return errors.New("moein must have currency options")
			}
			if docItem.Bedehkar == 0 {
				if docItem.Bestankar != docItem.CurrPrice*docItem.CurrRate {
					return errors.New("value in bestankar doesnt match with currency")
				}
			} else {
				if docItem.Bedehkar != docItem.CurrPrice*docItem.CurrRate {
					return errors.New("value in bedehkar doesnt match with currency")
				}
			}
		} else {
			if docItem.CurrPrice != 0 || docItem.Curr != "" || docItem.CurrRate != 0 {
				return errors.New("moein must not have currency options")
			}
		}
		return nil
	})
	return err
}

func (service *docService) FilterByMinorNum(minorNum string) ([]entity.Doc, error) {
	var docs []entity.Doc
	res := service.db.Model(&entity.Doc{}).Where("minor_num = ?", minorNum).Omit("DocItems").Order("id asc").Find(&docs)
	if res.Error != nil {
		return docs, res.Error
	}
	return docs, nil
}

func (service *docService) Numbering() error {
	err := service.DoInTransaction(func(tx *gorm.DB) error {
		var docs []entity.Doc
		res := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Model(&entity.Doc{}).Omit("DocItems").Order("year asc, month asc, day asc, hour asc, minute asc, second asc, atf_num asc").Find(&docs)
		if res.Error != nil {
			return res.Error
		}
		for i := range docs {
			docs[i].DocNum = i + 1
		}
		res = tx.Save(&docs)
		return res.Error
	})
	return err
}
