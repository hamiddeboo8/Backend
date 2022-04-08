package controller

import (
	"AccountingDoc/Gin-Server/entity"
	"AccountingDoc/Gin-Server/service"

	"github.com/gin-gonic/gin"
)

type DocController interface {
	FindAll() ([]entity.Doc, error)
	Save(*gin.Context) (entity.Doc, error)
	SaveByID(int64, *gin.Context) (entity.Doc, error)
	FindByID(id int64) (entity.Doc, error)
	FindDraftByID(id int64) (entity.DocEaseJson, error)
	ChangeState(id int64) (entity.Doc, error)
	CanEdit(id int64) bool
	CanEditDraft(id int64) bool
	CreateDraftDoc() (entity.DocEaseJson, error)
	FindDrafts() ([]entity.DocEaseJson, error)
	SaveDraft(id int64, c *gin.Context) (entity.DocEaseJson, error)
	RemoveDraft(id int64) (entity.DocEaseJson, error)
}

type controller struct {
	service service.DocService
}

func New(service service.DocService) DocController {
	return &controller{
		service: service,
	}
}

func (c *controller) FindAll() ([]entity.Doc, error) {
	var docs []entity.Doc
	docs, err := c.service.FindAll()
	return docs, err
}

func (c *controller) FindByID(id int64) (entity.Doc, error) {
	var doc entity.Doc
	doc, err := c.service.FindByID(id)
	return doc, err
}

func (c *controller) FindDraftByID(id int64) (entity.DocEaseJson, error) {
	var doc entity.DocDraft
	doc, err := c.service.FindDraftByID(id)
	if err != nil {
		return entity.DocEaseJson{}, err
	}
	return c.service.ConvertDraftToDoc(doc), err
}

func (c *controller) Save(ctx *gin.Context) (entity.Doc, error) {
	var doc entity.Doc
	err := ctx.ShouldBindJSON(&doc)
	if err != nil {
		return doc, err
	}
	_, err = c.service.Save(doc)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (c *controller) SaveByID(id int64, ctx *gin.Context) (entity.Doc, error) {
	var doc entity.Doc
	//fmt.Println(ctx.Request.Body)
	//fmt.Println(ctx.Request.Header)
	err := ctx.ShouldBindJSON(&doc)
	if err != nil {
		return doc, err
	}
	_, err = c.service.SaveByID(id, doc)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (c *controller) ChangeState(id int64) (entity.Doc, error) {
	doc, err := c.service.ChangeState(id)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (c *controller) CanEdit(id int64) bool {
	return c.service.CanEdit(id)
}

func (c *controller) CanEditDraft(id int64) bool {
	return c.service.CanEditDraft(id)
}

func (c *controller) CreateDraftDoc() (entity.DocEaseJson, error) {
	docDraft, err := c.service.CreateDraftDoc()
	if err != nil {
		return entity.DocEaseJson{}, err
	}
	return c.service.ConvertDraftToDoc(docDraft), nil
}

func (c *controller) FindDrafts() ([]entity.DocEaseJson, error) {
	var docs []entity.DocDraft
	docs, err := c.service.FindDrafts()
	if err != nil {
		return make([]entity.DocEaseJson, 0), err
	}
	var resDocs []entity.DocEaseJson
	for _, doc := range docs {
		resDocs = append(resDocs, c.service.ConvertDraftToDoc(doc))
	}
	return resDocs, err
}

func (c *controller) SaveDraft(id int64, ctx *gin.Context) (entity.DocEaseJson, error) {
	var doc entity.DocEaseJson
	err := ctx.ShouldBindJSON(&doc)
	if err != nil {
		//fmt.Println("xxxxxxxxxxxxxxxxxxxxxx\n")
		return doc, err
	}
	_, err = c.service.SaveDraftByID(id, doc)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (c *controller) RemoveDraft(id int64) (entity.DocEaseJson, error) {
	_, err := c.service.RemoveDraft(id)
	if err != nil {
		return entity.DocEaseJson{}, err
	}
	return entity.DocEaseJson{}, nil
}
