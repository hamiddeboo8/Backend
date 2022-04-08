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
	Save(doc entity.Doc) (entity.Doc, error)
	SaveByID(id uint64, doc entity.Doc) (entity.Doc, error)
	FindByID(id uint64) (entity.Doc, error)
	FindDraftByID(id uint64) (entity.DocDraft, error)
	ChangeState(id uint64) (entity.Doc, error)
	CanEdit(id uint64) bool
	CanEditDraft(id uint64) bool
	CreateDraftDoc() (entity.DocDraft, error)
	FindDrafts() ([]entity.DocDraft, error)
	SaveDraftByID(uint64, entity.DocDraft) (entity.DocDraft, error)
	RemoveDraft(id uint64) error
	//ConvertDocToDraft(docD entity.DocDraft) entity.DocDraft
	//ConvertDraftToDoc(docD entity.DocDraft) entity.DocDraft
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

func (service *docService) FindDrafts() ([]entity.DocDraft, error) {
	var docs []entity.DocDraft
	docs, err := service.docRepository.FindAllDraft()
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

func (service *docService) FindDraftByID(id uint64) (entity.DocDraft, error) {
	var docDraft entity.DocDraft
	docDraft, err := service.docRepository.FindDraftByID(id)
	//index := binarySearchDraft(service.draftDocs, 0, len(service.draftDocs), id)
	if err == nil {
		return docDraft, nil
	}
	return docDraft, errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
}

func (service *docService) Save(doc entity.Doc) (entity.Doc, error) {
	_, err := service.docRepository.FindByID(doc.ID)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), doc.ID)
	if err == nil {
		return entity.Doc{}, errors.New("doc id is not unique")
	}
	docDraft, err := service.docRepository.FindDraftByID(doc.ID)
	//index = binarySearchDraft(service.draftDocs, 0, len(service.draftDocs), doc.ID)
	if err != nil {
		return entity.Doc{}, errors.New("not in drafts")
	} else {
		err = service.docRepository.DeleteDraft(docDraft)
		if err != nil {
			return entity.Doc{}, errors.New("cannot delete from drafts")
		}
		glb, err := service.docRepository.FindGlobal()
		if err != nil {
			return entity.Doc{}, errors.New("cannot get total atf num")
		}
		doc.AtfNum = glb.AtfNumGlobal
		doc.DailyNum = glb.TodayCount
		glb.TodayCount += 1
		glb.AtfNumGlobal += 1
		glb, err = service.docRepository.UpdateGlobal(glb)
		if err != nil {
			return entity.Doc{}, errors.New("cannot update total atf num")
		}
		//service.DocIDsUpdate(len(service.docs) - 1)
		doc, err = service.docRepository.Save(doc)
		return doc, err
	}
}

func (service *docService) SaveByID(id uint64, doc entity.Doc) (entity.Doc, error) {
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

func (service *docService) ChangeState(id uint64) (entity.Doc, error) {
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

func (service *docService) CanEdit(id uint64) bool {
	doc, err := service.docRepository.FindByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	if err != nil {
		return false
	}
	if doc.State == "دائمی" {
		return false
	}
	return true
}

func (service *docService) CanEditDraft(id uint64) bool {
	//multi client
	_, err := service.docRepository.FindDraftByID(id)
	//index := binarySearchDoc(service.docs, 0, len(service.docs), id)
	return err == nil
}

func (service *docService) CreateDraftDoc() (entity.DocDraft, error) {
	doc, err := service.createNewDocDraft()
	if err != nil {
		return entity.DocDraft{}, err
	}
	return doc, nil
}

func (service *docService) createNewDocDraft() (entity.DocDraft, error) {
	var doc entity.DocDraft
	//geogDate := time.Now()
	//hirjiDate, _ := hijri.CreateHijriDate(geogDate, hijri.Default)
	pt := ptime.Now()
	doc.Year = pt.Year()
	doc.Month = int(pt.Month())
	doc.Day = pt.Day()
	doc.Hour = pt.Hour()
	doc.Minute = pt.Minute()
	doc.Second = pt.Second()
	doc.AtfNum = -1
	doc.DocNum = 0
	doc.MinorNum = ""
	doc.Desc = ""
	doc.State = "موقت"
	doc.DailyNum = -1
	doc.DocType = "عمومی"
	doc.EmitSystem = "سیستم حسابداری"
	doc.DocItems = make([]entity.DocItemDraft, 0)
	doc, err := service.docRepository.SaveDraft(doc)
	return doc, err
}

func (service *docService) SaveDraftByID(id uint64, doc entity.DocDraft) (entity.DocDraft, error) {
	doc_temp, err := service.docRepository.FindDraftByID(id)
	//index := binarySearchDraft(service.draftDocs, 0, len(service.draftDocs), id)
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
	//index := binarySearchDraft(service.draftDocs, 0, len(service.draftDocs), id)
	if err != nil {
		return errors.New("DocDraft with id " + strconv.FormatUint(id, 10) + " Not Found")
	}
	//service.draftDocs = append(service.draftDocs[:index], service.draftDocs[index+1:]...)
	return nil
}

/*func compareDate(first entity.Date, second entity.Date) int {
	if first.Year > second.Year {
		return 1
	}
	if first.Year == second.Year {
		if first.Month > second.Month {
			return 1
		}
		if first.Month == second.Month {
			if first.Day > second.Day {
				return 1
			}
			if first.Day == second.Day {
				if first.Hour > second.Hour {
					return 1
				}
				if first.Hour == second.Hour {
					if first.Minute > second.Minute {
						return 1
					}
					if first.Minute == second.Minute {
						if first.Second > second.Second {
							return 1
						}
						if first.Second == second.Second {
							return 0
						}
					}
				}
			}
		}
	}
	return -1
}*/
/*func (service *docService) DocIDsUpdate(index int) {
	doc := service.docs[index]
	temp := entity.Date{-1, -1, -1, -1, -1, -1}
	tempAtf := -1
	tempID := -1
	fmt.Printf("x0\n")
	for i, doc_ := range service.docs {
		fmt.Printf("x1: %d\n", i)
		if i == index {
			continue
		}
		if compareDate(doc_.DocDate, doc.DocDate) == 1 {
			fmt.Printf("x2: %d\n", i)
			if temp.Year == -1 || compareDate(doc_.DocDate, temp) == -1 || (compareDate(doc_.DocDate, temp) == 0 && doc_.AtfNum < tempAtf) {
				fmt.Printf("x3: %d\n", i)
				temp = doc_.DocDate
				tempAtf = doc_.AtfNum
				tempID = doc_.DocID
			}
		}
	}
	if tempID == -1 {
		fmt.Printf("x4: %d\n", len(service.docs))
		service.docs[index].DocID = len(service.docs)
	} else {
		fmt.Printf("x5: %d\n", tempID)
		service.docs[index].DocID = tempID
		for i, doc_ := range service.docs {
			if i == index {
				continue
			}
			if doc_.DocID >= tempID {
				service.docs[i].DocID += 1
			}
		}
	}
}*/
/*func (service *docService) ConvertDraftToDoc(docD entity.DocDraft) entity.DocEaseJson {
	var doc entity.DocEaseJson
	doc.AtfNum = *docD.AtfNum
	doc.DailyNum = *docD.DocID
	doc.Desc = docD.Desc
	doc.DocDate = docD.DocDate
	doc.DocID = *docD.DocID
	doc.DocItems = docD.DocItems
	doc.DocType = docD.DocType
	doc.EmitSystem = docD.EmitSystem
	doc.ID = docD.ID
	doc.MinorNum = docD.MinorNum
	doc.State = docD.State
	return doc
}

func (service *docService) ConvertDocToDraft(docD entity.DocEaseJson) entity.DocDraft {
	var doc entity.DocDraft
	doc.AtfNum = &service.globalVars.atfNumGlobal
	doc.DailyNum = &service.globalVars.todayCount
	doc.DocID = doc.AtfNum
	doc.Desc = docD.Desc
	doc.DocDate = docD.DocDate
	doc.DocItems = docD.DocItems
	doc.DocType = docD.DocType
	doc.EmitSystem = docD.EmitSystem
	doc.ID = docD.ID
	doc.MinorNum = docD.MinorNum
	doc.State = docD.State
	return doc
}
*/
/*func binarySearchDoc(arr []entity.Doc, st int, en int, id int64) int {
	if en == st {
		return -1
	}
	if en-st == 1 {
		if arr[st].ID == id {
			return st
		}
		return -1
	} else {
		mid := (st + en) / 2
		if arr[mid].ID == id {
			return mid
		} else if id < arr[mid].ID {
			return binarySearchDoc(arr, st, mid, id)
		} else {
			return binarySearchDoc(arr, mid+1, en, id)
		}
	}
}
func binarySearchDraft(arr []entity.DocDraft, st int, en int, id int64) int {
	if en == st {
		return -1
	}
	if en-st == 1 {
		if arr[st].ID == id {
			return st
		}
		return -1
	} else {
		mid := (st + en) / 2
		if arr[mid].ID == id {
			return mid
		} else if id < arr[mid].ID {
			return binarySearchDraft(arr, st, mid, id)
		} else {
			return binarySearchDraft(arr, mid+1, en, id)
		}
	}
}
*/
