package main

import (
	"AccountingDoc/Gin-Server/controller"
	"AccountingDoc/Gin-Server/repository"
	"AccountingDoc/Gin-Server/service"
	"strconv"

	"fmt"

	"github.com/gin-gonic/gin"
)

func CORS(c *gin.Context) {

	// First, we add the headers with need to enable CORS
	// Make sure to adjust these headers to your needs
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	// Second, we handle the OPTIONS problem
	if c.Request.Method != "OPTIONS" {

		c.Next()

	} else {
		fmt.Println(c.Request.Body)
		// Everytime we receive an OPTIONS request,
		// we just return an HTTP 200 Status Code
		// Like this, Angular can now do the real
		// request using any other method than OPTIONS
		c.AbortWithStatus(200)
	}
}

var (
	docRepository repository.DocRepository = repository.NewDocRepository()
	docService    service.DocService       = service.New(docRepository)
	DocController controller.DocController = controller.New(docService)
)

func main() {

	defer docRepository.CloseDB()

	r := gin.Default()

	r.Use(CORS)

	r.POST("/docs", func(c *gin.Context) {
		res, err := DocController.Save(c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs", func(c *gin.Context) {
		res, err := DocController.FindAll()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.FindByID(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.POST("/docs/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.SaveByID(id, c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})
	r.PUT("/docs/change_state/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.ChangeState(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/edit/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res := DocController.CanEdit(id)
		if res {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": res,
			})
		}
	})

	r.GET("/docs/drafts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.FindDraftByID(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/drafts/edit/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res := DocController.CanEditDraft(id)
		if res {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": res,
			})
		}
	})

	r.GET("/docs/drafts", func(c *gin.Context) {
		res, err := DocController.FindDrafts()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/drafts/create", func(c *gin.Context) {
		res, err := DocController.CreateDraftDoc()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.PUT("/docs/drafts/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.SaveDraft(id, c)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.PUT("/docs/drafts/delete/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res, err := DocController.RemoveDraft(id)
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/moeins", func(c *gin.Context) {
		res, err := DocController.FindMoeins()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/tafsilis", func(c *gin.Context) {
		res, err := DocController.FindTafsilis()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
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
