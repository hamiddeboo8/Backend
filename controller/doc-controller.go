package controller

import (
	"AccountingDoc/Gin-Server/entity"
	"AccountingDoc/Gin-Server/service"

	"github.com/gin-gonic/gin"
)

type DocController interface {
	FindAll() ([]entity.Doc, error)
	Save(*gin.Context) (entity.Doc, error)
	SaveByID(uint64, *gin.Context) (entity.Doc, error)
	FindByID(id uint64) (entity.Doc, error)
	FindDraftByID(id uint64) (entity.DocDraft, error)
	ChangeState(id uint64) (entity.Doc, error)
	CanEdit(id uint64) bool
	CanEditDraft(id uint64) bool
	CreateDraftDoc() (entity.DocDraft, error)
	FindDrafts() ([]entity.DocDraft, error)
	SaveDraft(id uint64, c *gin.Context) (entity.DocDraft, error)
	RemoveDraft(id uint64) (entity.DocDraft, error)
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

func (c *controller) FindByID(id uint64) (entity.Doc, error) {
	var doc entity.Doc
	doc, err := c.service.FindByID(id)
	return doc, err
}

func (c *controller) FindDraftByID(id uint64) (entity.DocDraft, error) {
	var doc entity.DocDraft
	doc, err := c.service.FindDraftByID(id)
	if err != nil {
		return entity.DocDraft{}, err
	}
	return doc, err
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

func (c *controller) SaveByID(id uint64, ctx *gin.Context) (entity.Doc, error) {
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

func (c *controller) ChangeState(id uint64) (entity.Doc, error) {
	doc, err := c.service.ChangeState(id)
	if err != nil {
		return doc, err
	}
	return doc, nil
}

func (c *controller) CanEdit(id uint64) bool {
	return c.service.CanEdit(id)
}

func (c *controller) CanEditDraft(id uint64) bool {
	return c.service.CanEditDraft(id)
}

func (c *controller) CreateDraftDoc() (entity.DocDraft, error) {
	docDraft, err := c.service.CreateDraftDoc()
	if err != nil {
		return entity.DocDraft{}, err
	}
	return docDraft, nil
}

func (c *controller) FindDrafts() ([]entity.DocDraft, error) {
	var docs []entity.DocDraft
	docs, err := c.service.FindDrafts()
	if err != nil {
		return make([]entity.DocDraft, 0), err
	}
	return docs, err
}

func (c *controller) SaveDraft(id uint64, ctx *gin.Context) (entity.DocDraft, error) {
	var doc entity.DocDraft
	err := ctx.ShouldBindJSON(&doc)
	if err != nil {
		//fmt.Println("xxxxxxxxxxxxxxxxxxxxxx\n")
		return doc, err
	}
	_, err = c.service.SaveDraftByID(id, doc)
	if err != nil {
		//fmt.Println("yyyyyyyyyyyyyyyyyyyyyyy\n")
		return doc, err
	}
	return doc, nil
}

func (c *controller) RemoveDraft(id uint64) (entity.DocDraft, error) {
	err := c.service.RemoveDraft(id)
	return entity.DocDraft{}, err
}
