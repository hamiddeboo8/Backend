package main

import (
	"AccountingDoc/Gin-Server/entity"
	"AccountingDoc/Gin-Server/service"
	"strconv"

	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	db         *gorm.DB           = service.NewDbConnection()
	DocService service.DocService = service.New(db)
)

func main() {

	defer DocService.CloseDB()

	r := gin.Default()

	r.Use(CORS)

	r.POST("/docs", func(c *gin.Context) {
		var doc entity.Doc
		err := c.ShouldBindJSON(&doc)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err = DocService.Save(doc)
			//res, err := DocController.Save(c)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Successfully Saved",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs", func(c *gin.Context) {
		res, err := DocService.FindAll()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			res, err := DocService.FindByID(id)
			if err == nil {
				c.JSON(200, res)
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.PUT("/docs/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			//race condition
			var doc entity.AddRemoveDocItem
			err := c.ShouldBindJSON(&doc)
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			} else {
				err = DocService.SaveByID(id, doc)
				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})
				} else {
					c.JSON(200, gin.H{
						"message": "Successfully Updated",
					})
				}
			}
		}
	})
	r.PUT("/docs/change_state/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.ChangeState(id)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Successfully Updated",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs/edit/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.CanEdit(id)
			if err != nil {
				c.JSON(200, gin.H{
					"message": "Can be edited",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs/drafts/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			res, err := DocService.FindDraftByID(id)
			if err == nil {
				c.JSON(200, res)
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs/drafts/edit/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.CanEditDraft(id)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Can be edited",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs/drafts", func(c *gin.Context) {
		res, err := DocService.FindDrafts()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/drafts/create", func(c *gin.Context) {
		//because some initializes things like date or (if implement => atf and daily number)
		res, err := DocService.CreateDraftDoc()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.PUT("/docs/drafts/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			var doc entity.AddRemoveDocDraftItem
			err := c.ShouldBindJSON(&doc)
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			} else {
				err = DocService.SaveDraftByID(id, doc)
				if err != nil {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})
				}
				if err == nil {
					c.JSON(200, gin.H{
						"message": "Successfully Updated",
					})
				} else {
					c.JSON(500, gin.H{
						"message": err.Error(),
					})
				}
			}
		}
	})

	r.PUT("/docs/drafts/delete/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.RemoveDraft(id)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Successfully Deleted",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	r.GET("/docs/moeins", func(c *gin.Context) {
		res, err := DocService.FindMoeins()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.GET("/docs/tafsilis", func(c *gin.Context) {
		res, err := DocService.FindTafsilis()
		if err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
	})

	r.Run(":5710")
}
