package main

import (
	"AccountingDoc/Gin-Server/controller"
	"AccountingDoc/Gin-Server/service"
	"strconv"

	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	docService    service.DocService       = service.New()
	DocController controller.DocController = controller.New(docService)
)

func main() {

	r := gin.Default()

	r.Use(CORS)

	r.POST("/docs", func(c *gin.Context) {
		res, err := DocController.Save(c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.GET("/docs", func(c *gin.Context) {
		res, err := DocController.FindAll()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.GET("/docs/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.FindByID(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.POST("/docs/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.SaveByID(id, c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})
	r.PUT("/docs/change_state/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.ChangeState(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.GET("/docs/edit/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res := DocController.CanEdit(id)
		if res {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", res))
		}
	})

	r.GET("/docs/drafts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.FindDraftByID(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.GET("/docs/drafts/edit/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res := DocController.CanEditDraft(id)
		if res {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", res))
		}
	})

	r.GET("/docs/drafts", func(c *gin.Context) {
		res, err := DocController.FindDrafts()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.GET("/docs/drafts/create", func(c *gin.Context) {
		res, err := DocController.CreateDraftDoc()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.PUT("/docs/drafts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.SaveDraft(id, c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	r.PUT("/docs/drafts/delete/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		res, err := DocController.RemoveDraft(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})

	/*r.GET("/docs/create", func(c *gin.Context) {
		res, err := DocController.NextDocInitial()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, fmt.Errorf("%v", err))
		}
	})*/

	r.Run(":5710")
}
