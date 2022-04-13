package main

import (
	"AccountingDoc/Gin-Server/entity"
	"AccountingDoc/Gin-Server/service"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	mu               sync.RWMutex
	atf_num_global   int
	daily_num_global int
)

func Test_GetPostDocs(t *testing.T) {
	db := service.NewDbConnection()
	DocService := service.New(db)
	defer DocService.CloseDB()

	n := 20

	//errs := make(chan error)
	results := make(chan struct{}, n)
	var expected []entity.Doc

	moeins := createRandomMoeins()
	tafsilis := createRandomTafsilis()

	atf_num_global = 1
	daily_num_global = 1
	minorNums := []string{"", "0120", "0310", "0010", "5000", "3050"}
	descs := []string{"desc1", "desc2", "desc3", "desc4", "desc5", "desc6", "desc7", "desc8"}
	for i := 0; i < n; i++ {
		go func() {
			doc, err := insertRandomTestDoc(DocService, moeins, tafsilis, minorNums, descs)
			if err != nil {
				t.Errorf(err.Error())
			}

			//errs <- err
			expected = append(expected, doc)
			results <- struct{}{}
		}()
	}

	for i := 0; i < n; i++ {
		<-results
	}

	req, w := SetupGetDocs(DocService, "/docs")
	if req.Method != http.MethodGet {
		t.Errorf("HTTP request method error")
	}
	if w.Code != http.StatusOK {
		t.Errorf("HTTP request status code error")
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}

	actual := []entity.Doc{}
	if err := json.Unmarshal(body, &actual); err != nil {
		t.Errorf(err.Error())
	}

	for i := 0; i < n; i++ {
		if !equalDoc(&actual[i], &expected[actual[i].AtfNum-1]) {
			t.Errorf("atf number ", i+1, " is not correct")
		}
	}
	ClearTable(DocService)
}

func insertRandomTestDoc(DocService service.DocService, moeins []entity.Moein, tafsilis []entity.Tafsili, minorNums []string, descs []string) (entity.Doc, error) {
	mu.RLock()
	defer mu.RUnlock()
	var doc entity.Doc
	num_of_items := rand.Intn(100)
	var docItems []entity.DocItem
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, createRandomDocItem(moeins, tafsilis))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	atf_num_global += 1
	doc.DailyNum = daily_num_global
	daily_num_global += 1
	doc.Year = 1401
	doc.Month = rand.Intn(12) + 1
	doc.Day = rand.Intn(29) + 1
	doc.Hour = rand.Intn(23) + 1
	doc.Minute = rand.Intn(59) + 1
	doc.Second = rand.Intn(59) + 1
	doc.Desc = descs[rand.Intn(8)]
	doc.DocNum = 0
	doc.DocType = "عمومی"
	doc.EmitSystem = "سیستم حسابداری"
	doc.IsChanging = false
	doc.MinorNum = minorNums[rand.Intn(6)]
	doc.State = "موقت"
}

func equalDoc(doc1 *entity.Doc, doc2 *entity.Doc) bool {
	if doc1.AtfNum != doc2.AtfNum {
		return false
	}
	if doc1.DailyNum != doc2.DailyNum {
		return false
	}
	if doc1.Year != doc2.Year || doc1.Month != doc2.Month || doc1.Day != doc2.Day {
		return false
	}
	if doc1.Hour != doc2.Hour || doc1.Minute != doc2.Minute || doc1.Second != doc2.Second {
		return false
	}
	if doc1.Desc != doc2.Desc {
		return false
	}
	if doc1.DocNum != doc2.DocNum {
		return false
	}
	if doc1.DocType != doc2.DocType {
		return false
	}
	if doc1.EmitSystem != doc2.EmitSystem {
		return false
	}
	if doc1.IsChanging != doc2.IsChanging {
		return false
	}
	if doc1.State != doc2.State {
		return false
	}
	if doc1.MinorNum != doc2.MinorNum {
		return false
	}
	for i := range doc1.DocItems {
		if !equalDocItem(&doc1.DocItems[i], &doc2.DocItems[i]) {
			return false
		}
	}
	return true
}

func equalDocItem(item1 *entity.DocItem, item2 *entity.DocItem) bool {
	if item1.Bedehkar != item2.Bedehkar {
		return false
	}
	if item1.Bestankar != item2.Bestankar {
		return false
	}
	if item1.Curr != item2.Curr {
		return false
	}
	if item1.CurrPrice != item2.CurrPrice {
		return false
	}
	if item1.CurrRate != item2.CurrRate {
		return false
	}
	if item1.Desc != item2.Desc {
		return false
	}
	if item1.Num != item2.Num {
		return false
	}
	if !equalCode(&item1.Moein, &item2.Moein) {
		return false
	}
	if !equalCode(&item1.Tafsili, &item2.Tafsili) {
		return false
	}
	return true
}

func equalCode(i1 interface{}, i2 interface{}) bool {
	code1, ok := i1.(*entity.Code)
	if !ok {
		return false
	}
	code2, ok := i2.(*entity.Code)
	if !ok {
		return false
	}
	if code1.CodeVal != code2.CodeVal {
		return false
	}
	if code1.CurrPossible != code2.CurrPossible {
		return false
	}
	if code1.Name != code2.Name {
		return false
	}
	if code1.TrackPossible != code2.TrackPossible {
		return false
	}
	return true
}

func ClearTable(DocService service.DocService) {
	DocService.GetDB().Exec("DELETE FROM docs")
	DocService.GetDB().Exec("ALTER SEQUENCE docs_id_seq RESTART WITH 1")
}

func SetupGetDocs(DocService service.DocService, url string) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()

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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}
