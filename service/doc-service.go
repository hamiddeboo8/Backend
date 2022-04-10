package service

import (
	"AccountingDoc/Gin-Server/entity"
	"AccountingDoc/Gin-Server/repository"
	"errors"
	"strconv"

	ptime "github.com/yaa110/go-persian-calendar"
)

type DocService interface {
	FindAll() ([]entity.Doc, error)
	Save(doc entity.Doc) error
	SaveByID(id uint64, doc entity.Doc) error
	FindByID(id uint64) (entity.Doc, error)
	ChangeState(id uint64) error
	CanEdit(id uint64) error

	FindDraftByID(id uint64) (entity.DocDraft, error)
	CanEditDraft(id uint64) error
	CreateDraftDoc() (entity.DocDraft, error)
	FindDrafts() ([]entity.DocDraft, error)
	SaveDraftByID(uint64, entity.DocDraft) error
	RemoveDraft(id uint64) error

	FindMoeins() ([]entity.Moein, error)
	FindTafsilis() ([]entity.Tafsili, error)
}

type docService struct {
	docRepository repository.DocRepository
}

func New(repo repository.DocRepository) DocService {
	return &docService{
		docRepository: repo,
	}
}

func (service *docService) FindAll() ([]entity.Doc, error) {
	var docs []entity.Doc
	docs, err := service.docRepository.FindAll()
	if err != nil {
		return docs, err
	}
	return docs, nil
}

func (service *docService) FindByID(id uint64) (entity.Doc, error) {
	var doc entity.Doc
	doc, err := service.docRepository.FindByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	if err == nil {
		return doc, nil
	}
	return doc, errors.New("Doc with id " + strconv.FormatUint(id, 10) + " Not Found")
}

func (service *docService) FindDrafts() ([]entity.DocDraft, error) {
	var docs []entity.DocDraft
	docs, err := service.docRepository.FindAllDraft()
	if err != nil {
		return docs, err
	}
	return docs, nil
}

func (service *docService) FindDraftByID(id uint64) (entity.DocDraft, error) {
	var docDraft entity.DocDraft
	docDraft, err := service.docRepository.FindDraftByID(id)
	//index := binarySearchDraft(service.draftDocs, 0, len(service.draftDocs), id)
	if err == nil {
		return docDraft, nil
	}
	return docDraft, errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
}

func (service *docService) Save(doc entity.Doc) error {
	docDraft, err := service.docRepository.FindDraftByID(doc.ID)
	if err != nil {
		return errors.New("not in drafts")
	} else {
		err = service.docRepository.DeleteDraft(docDraft)
		if err != nil {
			return errors.New("cannot delete from drafts")
		}
		//handle lock on database
		glb, err := service.docRepository.FindGlobal()
		if err != nil {
			return errors.New("cannot get total atf num")
		}
		doc.AtfNum = glb.AtfNumGlobal
		doc.DailyNum = glb.TodayCount
		glb.TodayCount += 1
		glb.AtfNumGlobal += 1
		glb, err = service.docRepository.UpdateGlobal(glb)
		//
		if err != nil {
			return errors.New("cannot update total atf num")
		}
		err = service.docRepository.Save(doc)
		return err
	}
}

func (service *docService) SaveByID(id uint64, doc entity.Doc) error {
	doc_temp, err := service.docRepository.FindByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	if err != nil {
		return doc_temp, errors.New("Doc with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	doc, err = service.docRepository.Update(doc)
	if err != nil {
		return doc_temp, errors.New("Doc with id " + strconv.FormatUint(id, 10) + " cannot update")
	}
	//service.DocIDsUpdate(len(service.docs) - 1)
	return doc, nil
}

func (service *docService) ChangeState(id uint64) error {
	doc, err := service.docRepository.FindByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	if err != nil {
		return doc, errors.New("Doc with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	if doc.State == "دائمی" {
		return doc, errors.New("doc state is permanent")
	}
	doc.State = "دائمی"
	doc, err = service.docRepository.Update(doc)
	return doc, err
}

func (service *docService) CanEdit(id uint64) error {
	doc, err := service.docRepository.FindByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	if err != nil {
		return false
	}
	if doc.State == "دائمی" {
		return false
	}
	if doc.IsChanging {
		return false
	}
	doc.IsChanging = true
	doc, err = service.docRepository.Update(doc)
	return err == nil
}

func (service *docService) CanEditDraft(id uint64) error {
	//multi client
	_, err := service.docRepository.FindDraftByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	return err == nil
}

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
	doc, err := service.docRepository.SaveDraft(doc)
	if err != nil {
		return entity.DocDraft{}, err
	}
	return doc, nil
}

func (service *docService) SaveDraftByID(id uint64, doc entity.DocDraft) error {
	doc_temp, err := service.docRepository.FindDraftByID(id)
	if err != nil {
		return doc_temp, errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	doc, err = service.docRepository.UpdateDraft(doc)
	if err != nil {
		return doc_temp, errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " cannot update")
	}
	return doc, nil
}

func (service *docService) RemoveDraft(id uint64) error {
	err := service.docRepository.DeleteDraftByID(id)
	if err != nil {
		return errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	return nil
}

func (service *docService) FindMoeins() ([]entity.Moein, error) {
	var codes []entity.Moein
	codes, err := service.docRepository.FindMoeins()
	if err != nil {
		return codes, err
	}
	return codes, nil
}
func (service *docService) FindTafsilis() ([]entity.Tafsili, error) {
	var codes []entity.Tafsili
	codes, err := service.docRepository.FindTafsilis()
	if err != nil {
		return codes, err
	}
	return codes, nil
}
