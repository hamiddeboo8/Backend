package service

import (
	"AccountingDoc/Gin-Server/entity"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AnsReq struct {
	Message string `json:"message"`
}

var (
	mu               sync.RWMutex
	mu2              []sync.RWMutex
	atf_num_global   int
	daily_num_global int
	moeins           []entity.Moein
	tafsilis         []entity.Tafsili
	minorNums        []string
	descs            []string
	currs            []string
	currRates        []int
	descItems        []string
)

func NewDbConnectionTest() *gorm.DB {
	dsn := "host=localhost user=postgres password=123581321345589144Hamidreza. dbname=AccountingDocsTest port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error in openning Database")
	}

	db.AutoMigrate(&entity.GlobalVars{}, &entity.Doc{}, &entity.DocItem{})

	// if db.Migrator().HasTable(&entity.GlobalVars{}) {
	// 	db.Migrator().DropTable(&entity.GlobalVars{})
	// }
	// if db.Migrator().HasTable(&entity.Doc{}) {
	// 	db.Migrator().DropTable(&entity.Doc{})
	// }
	// if db.Migrator().HasTable(&entity.DocItem{}) {
	// 	db.Migrator().DropTable(&entity.DocItem{})
	// }

	if !db.Migrator().HasTable(&entity.Moein{}) {
		db.Migrator().CreateTable(&entity.Moein{})
	}
	if !db.Migrator().HasTable(&entity.Tafsili{}) {
		db.Migrator().CreateTable(&entity.Tafsili{})
	}
	if !db.Migrator().HasTable(&entity.GlobalVars{}) {
		db.Migrator().CreateTable(&entity.GlobalVars{})
	}
	if !db.Migrator().HasTable(&entity.Doc{}) {
		db.Migrator().CreateTable(&entity.Doc{})
	}
	if !db.Migrator().HasTable(&entity.DocItem{}) {
		db.Migrator().CreateTable(&entity.DocItem{})
	}
	return db
}

func GetMoeinTafsiliFromDB(DocService DocService) ([]entity.Moein, []entity.Tafsili, error) {
	var moeins []entity.Moein
	var tafsilis []entity.Tafsili
	res := DocService.GetDB().Find(&moeins)
	if res.Error != nil {
		return []entity.Moein{}, []entity.Tafsili{}, res.Error
	}
	res = DocService.GetDB().Find(&tafsilis)
	if res.Error != nil {
		return []entity.Moein{}, []entity.Tafsili{}, res.Error
	}
	return moeins, tafsilis, nil
}
func GetGlobalFromDB(DocService DocService) (int, int, error) {
	var glb entity.GlobalVars
	res := DocService.GetDB().Find(&glb)
	if res.Error != nil {
		return -1, -1, res.Error
	}
	return glb.AtfNumGlobal, glb.TodayCount, nil
}

func CreateMoeinTafsili(DocService DocService) ([]entity.Moein, []entity.Tafsili, error) {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numberRunes := []rune("0123456789")
	moeins := createRandomMoeins(10, numberRunes, letterRunes)
	tafsilis := createRandomTafsilis(10, numberRunes, letterRunes)
	res := DocService.GetDB().Save(&moeins)
	if res.Error != nil {
		return []entity.Moein{}, []entity.Tafsili{}, res.Error
	}
	res = DocService.GetDB().Save(&tafsilis)
	if res.Error != nil {
		return []entity.Moein{}, []entity.Tafsili{}, res.Error
	}
	return moeins, tafsilis, nil
}

func Init(DocService DocService, n int) ([]entity.Doc, error) {
	results := make(chan struct{}, n)
	var docs = make([]entity.Doc, n)

	var err error
	//moeins, tafsilis, err = GetMoeinTafsiliFromDB(DocService)
	moeins, tafsilis, err = CreateMoeinTafsili(DocService)
	if err != nil {
		return []entity.Doc{}, err
	}

	// atf_num_global, daily_num_global, err = GetGlobalFromDB(DocService)
	// if err != nil {
	// 	return []entity.Doc{}, err
	// }
	atf_num_global = 1
	daily_num_global = 1

	minorNums = []string{"", "0120", "0310", "0010", "5000", "3050"}
	descs = []string{"desc1", "desc2", "desc3", "desc4", "desc5", "desc6", "desc7", "desc8"}
	currs = []string{"$", "€", "¥", "£", "₽"}
	currRates = []int{27800, 30200, 22200, 36400, 348}
	descItems = []string{}
	for i := 0; i < 50; i++ {
		descItems = append(descItems, "descItem"+strconv.Itoa(i+1))
	}
	for i := 0; i < n; i++ {
		go func() {
			doc, err := insertRandomTestDoc(DocService, moeins, tafsilis, minorNums, descs, descItems, currs, currRates)
			if err != nil {
				panic(err.Error())
			}

			docs = append(docs, doc)
			results <- struct{}{}
		}()
	}
	for i := 0; i < 5; i++ {
		doc, err := insertRandomPermanentTestDoc(DocService, moeins, tafsilis, minorNums, descs, descItems, currs, currRates, i+1)
		if err != nil {
			panic(err.Error())
		}
		docs = append(docs, doc)
	}

	for i := 0; i < n; i++ {
		<-results
	}
	return docs, nil
}
func Test_GetDocs(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 1000
	rand.Seed(int64(n))
	expected, err := Init(DocService, n)
	if err != nil {
		t.Errorf(err.Error())
	}

	req, w := SetupGetDocs(DocService)
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

	t.Log(len(actual), len(expected))
	for i := 0; i < n; i++ {
		if equalDoc(actual[i], expected[actual[i].AtfNum-1]) != "" {
			t.Errorf("atf number %d is not correct expected %d instead of %d", i+1, expected[i].AtfNum, actual[i].AtfNum)
		}
	}
}

func Test_PostDocs(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 10
	rand.Seed(int64(n))

	_, err := Init(DocService, n)
	if err != nil {
		t.Errorf(err.Error())
	}

	for i := 0; i < n; i++ {
		go func() {
			doc := createRandomTestDoc()
			reqBody, err := json.Marshal(doc)
			if err != nil {
				t.Errorf(err.Error())
			}
			req, w := SetupPostDocs(DocService, bytes.NewBuffer(reqBody))
			if req.Method != http.MethodPost {
				t.Errorf("HTTP request method error")
			}
			if w.Code != http.StatusOK {
				t.Errorf("HTTP request status code error")
			}
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error())
			}
			var show AnsReq
			if err := json.Unmarshal(body, &show); err != nil {
				t.Errorf(err.Error())
			}
			t.Log(show.Message)
		}()
	}

	var doc entity.Doc
	num_of_items := 10
	var docItems []entity.DocItem
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	//atf_num_global += 1
	doc.DailyNum = daily_num_global
	//daily_num_global += 1
	doc.Year = 1401
	doc.Month = 23
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
	reqBody, err := json.Marshal(doc)
	if err != nil {
		t.Errorf(err.Error())
	}
	req, w := SetupPostDocs(DocService, bytes.NewBuffer(reqBody))
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request method error")
	}
	if w.Code != 500 {
		t.Errorf("HTTP request status code error")
	}
	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}
	var show AnsReq
	if err := json.Unmarshal(body, &show); err != nil {
		t.Errorf(err.Error())
	}
	expected := "ERROR: new row for relation \"docs\" violates check constraint \"chk_docs_month\" (SQLSTATE 23514)"
	if show.Message != expected {
		t.Errorf("%s instead of %s", expected, show.Message)
	}
	////////////////////////////////////////////
	doc = entity.Doc{}
	num_of_items = 0
	docItems = []entity.DocItem{}
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	//atf_num_global += 1
	doc.DailyNum = daily_num_global
	//daily_num_global += 1
	doc.Year = 1401
	doc.Month = rand.Intn(11) + 2
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
	reqBody, err = json.Marshal(doc)
	if err != nil {
		t.Errorf(err.Error())
	}
	req, w = SetupPostDocs(DocService, bytes.NewBuffer(reqBody))
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request method error")
	}
	if w.Code != 500 {
		t.Errorf("HTTP request status code error")
	}
	body, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}
	show = AnsReq{}
	if err := json.Unmarshal(body, &show); err != nil {
		t.Errorf(err.Error())
	}
	expected = "must have at least one doc item"
	if show.Message != expected {
		t.Errorf("%s instead of %s", expected, show.Message)
	}
	////////////////////////////////////////////
	doc = entity.Doc{}
	num_of_items = 10
	docItems = []entity.DocItem{}
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	//atf_num_global += 1
	doc.DailyNum = daily_num_global
	//daily_num_global += 1
	doc.Year = 1401
	doc.Month = 1
	doc.Day = 23
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
	reqBody, err = json.Marshal(doc)
	if err != nil {
		t.Errorf(err.Error())
	}
	req, w = SetupPostDocs(DocService, bytes.NewBuffer(reqBody))
	if req.Method != http.MethodPost {
		t.Errorf("HTTP request method error")
	}
	if w.Code != 500 {
		t.Errorf("HTTP request status code error")
	}
	body, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}
	show = AnsReq{}
	if err := json.Unmarshal(body, &show); err != nil {
		t.Errorf(err.Error())
	}
	expected = "date must be bigger than latest permanent doc"
	if show.Message != expected {
		t.Errorf("%s instead of %s", expected, show.Message)
	}
}

func Test_GetDocID(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 100
	rand.Seed(int64(n))
	docs, err := Init(DocService, n)
	if err != nil {
		t.Errorf(err.Error())
	}

	ch := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		go func() {
			idx := rand.Intn(n)
			req, w := SetupGetDoc(DocService, idx+1)
			if req.Method != http.MethodGet {
				t.Errorf("HTTP request method error")
			}
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error())
			}
			if w.Code != http.StatusOK {
				var actual AnsReq
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Errorf(err.Error())
				}
				t.Log(actual.Message)
				t.Errorf("HTTP request status code error")
			}
			var actual entity.Doc
			if err := json.Unmarshal(body, &actual); err != nil {
				t.Errorf(err.Error())
			}
			if equalDoc(actual, docs[idx]) != "" {
				t.Logf(equalDoc(actual, docs[idx]))
				t.Errorf("expected atf_num %d instead of %d", docs[idx].AtfNum, actual.AtfNum)
			}
			ch <- struct{}{}
		}()
	}
	for i := 0; i < n; i++ {
		<-ch
	}
}

func Test_EditDoc(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 10
	r := 10

	editReqs := make([]int, n)

	var docs = make([]entity.Doc, n)
	rand.Seed(int64(n))
	docs, err := Init(DocService, n)
	if err != nil {
		t.Errorf(err.Error())
	}

	t.Log(len(docs))
	ch := make(chan struct{}, r)
	mu2 = make([]sync.RWMutex, n)

	for i := 0; i < r; i++ {
		go func() {
			idx := rand.Intn(n)

			_, w := SetupCanEditDoc(DocService, idx+1)
			if w.Code == http.StatusOK {
				mu2[idx].Lock()
				if editReqs[idx] != 0 {
					mu2[idx].Unlock()
					t.Errorf("more than one changing doc with id %d", idx+1)
				}
				editReqs[idx] += 1
				mu2[idx].Unlock()
				doc_edit := createRandomEditDoc(docs[idx])
				_, w := SetupPutDoc(DocService, doc_edit)
				body, err := ioutil.ReadAll(w.Body)
				if err != nil {
					t.Errorf(err.Error())
				}
				if w.Code != http.StatusOK {
					var actual AnsReq
					if err := json.Unmarshal(body, &actual); err != nil {
						t.Errorf(err.Error())
					}
					t.Log(actual.Message)
					t.Errorf("HTTP Edit request status code error")
				} else {
					_, w := SetupIsChangeDoc(DocService, idx+1)
					if w.Code != http.StatusOK {
						var actual AnsReq
						if err := json.Unmarshal(body, &actual); err != nil {
							t.Errorf(err.Error())
						}
						t.Log(actual.Message)
						t.Errorf("HTTP After Edit request status code error")
					}
				}
			} else {
				mu2[idx].Lock()
				if editReqs[idx] != 1 {
					mu2[idx].Unlock()
					t.Errorf("problem in changing doc with id %d", idx+1)
				} else {
					editReqs[idx] -= 1
					mu2[idx].Unlock()
				}
			}
			ch <- struct{}{}
		}()
	}
	for i := 0; i < r; i++ {
		<-ch
	}
}

func SetupGetDocs(DocService DocService) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()
	url := "/docs"
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

func SetupIsChangeDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()
	url := "/docs/changing/" + strconv.FormatInt(int64(idx), 10)
	r.PUT("/docs/changing/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.ChangeIsChange(id)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Successfully Exit",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}
func SetupCanEditDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()
	url := "/docs/edit/" + strconv.FormatInt(int64(idx), 10)
	r.GET("/docs/edit/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.CanEdit(id)
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupGetDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()
	url := "/docs/" + strconv.FormatInt(int64(idx), 10)
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupPostDocs(DocService DocService, body *bytes.Buffer) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()

	r.POST("/docs", func(c *gin.Context) {
		var doc entity.Doc
		err := c.ShouldBindJSON(&doc)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err = DocService.Save(doc)
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
	req, err := http.NewRequest(http.MethodPost, "/docs", body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupPutDoc(DocService DocService, doc entity.AddRemoveDocItem) (*http.Request, *httptest.ResponseRecorder) {
	r := gin.Default()
	url := "/docs/" + strconv.FormatUint(uint64(doc.AtfNum), 10)
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
	reqBody, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func createRandomMoeins(n int, numberRunes []rune, letterRunes []rune) []entity.Moein {
	var moeins []entity.Moein
	for i := 0; i < n; i++ {
		moein := entity.Moein{}
		for {
			flag := 0
			moein.CodeVal = string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)])
			for j := 0; j < len(moeins); j++ {
				if moeins[j].CodeVal == moein.CodeVal {
					flag = 1
					break
				}
			}
			if flag == 0 {
				break
			}
		}
		moein.CurrPossible = (rand.Intn(2) == 0)
		moein.TrackPossible = (rand.Intn(10) != 0)
		moein.Name = "moein" + string(numberRunes[i])
		moeins = append(moeins, moein)
	}
	return moeins
}

func createRandomTafsilis(n int, numberRunes []rune, letterRunes []rune) []entity.Tafsili {
	var tafsilis []entity.Tafsili
	for i := 0; i < n; i++ {
		tafsili := entity.Tafsili{}
		for {
			flag := 0
			if i == 0 {
				tafsili.CodeVal = ""
			} else {
				tafsili.CodeVal = string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)]) + string(numberRunes[rand.Intn(10)])
			}
			for j := 0; j < len(tafsilis); j++ {
				if tafsilis[j].CodeVal == tafsili.CodeVal {
					flag = 1
					break
				}
			}
			if flag == 0 {
				break
			}
		}
		tafsili.Name = "tafsili" + string(numberRunes[i])
		tafsilis = append(tafsilis, tafsili)
	}
	return tafsilis
}

func insertRandomPermanentTestDoc(DocService DocService, moeins []entity.Moein, tafsilis []entity.Tafsili, minorNums []string, descs []string, descItems []string, currs []string, currRates []int, count int) (entity.Doc, error) {
	mu.Lock()
	defer mu.Unlock()
	var doc entity.Doc
	num_of_items := rand.Intn(99) + 1
	var docItems []entity.DocItem
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	atf_num_global += 1
	doc.DailyNum = daily_num_global
	daily_num_global += 1
	doc.Year = 1401
	doc.Month = 1
	doc.Day = count * 5
	doc.Hour = rand.Intn(23) + 1
	doc.Minute = rand.Intn(59) + 1
	doc.Second = rand.Intn(59) + 1
	doc.Desc = descs[rand.Intn(8)]
	doc.DocNum = 0
	doc.DocType = "عمومی"
	doc.EmitSystem = "سیستم حسابداری"
	doc.IsChanging = false
	doc.MinorNum = minorNums[rand.Intn(6)]
	doc.State = "دائمی"
	if err := DocService.Save(doc); err != nil {
		return doc, err
	}
	return doc, nil
}

func insertRandomTestDoc(DocService DocService, moeins []entity.Moein, tafsilis []entity.Tafsili, minorNums []string, descs []string, descItems []string, currs []string, currRates []int) (entity.Doc, error) {
	mu.Lock()
	defer mu.Unlock()
	var doc entity.Doc
	num_of_items := rand.Intn(99) + 1
	var docItems []entity.DocItem
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	atf_num_global += 1
	doc.DailyNum = daily_num_global
	daily_num_global += 1
	doc.Year = 1401
	doc.Month = rand.Intn(11) + 2
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
	if err := DocService.Save(doc); err != nil {
		return doc, err
	}
	return doc, nil
}

func createRandomTestDoc() entity.Doc {
	mu.Lock()
	defer mu.Unlock()
	var doc entity.Doc
	num_of_items := rand.Intn(99) + 1
	var docItems []entity.DocItem
	for i := 0; i < num_of_items; i++ {
		docItems = append(docItems, *createRandomDocItem(i + 1))
	}
	doc.DocItems = docItems
	doc.AtfNum = atf_num_global
	atf_num_global += 1
	doc.DailyNum = daily_num_global
	daily_num_global += 1
	doc.Year = 1401
	doc.Month = rand.Intn(11) + 2
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
	return doc
}
func createRandomEditDoc(doc entity.Doc) entity.AddRemoveDocItem {
	var doc_res entity.AddRemoveDocItem = entity.AddRemoveDocItem{ID: doc.ID, DocNum: doc.DocNum, Year: doc.Year, Month: doc.Month, Day: doc.Day, Hour: doc.Hour, Minute: doc.Minute, Second: doc.Second, AtfNum: doc.AtfNum, MinorNum: doc.MinorNum, Desc: doc.Desc, State: doc.State, DailyNum: doc.DailyNum, DocType: doc.DocType, EmitSystem: doc.EmitSystem}
	var editDefiner []bool
	for i := 0; i < 6; i++ {
		editDefiner = append(editDefiner, rand.Intn(2) == 0)
	}
	if editDefiner[0] {
		doc_res.Year = 1401
		doc_res.Month = rand.Intn(12) + 1
		doc_res.Day = rand.Intn(29) + 1
		doc_res.Hour = rand.Intn(23) + 1
		doc_res.Minute = rand.Intn(59) + 1
		doc_res.Second = rand.Intn(59) + 1
	}
	if editDefiner[1] {
		doc_res.MinorNum = minorNums[rand.Intn(6)]
	}
	if editDefiner[2] {
		doc_res.Desc = descs[rand.Intn(8)]
	}
	if editDefiner[3] {
		num_of_items := rand.Intn(99) + 1
		var docItems []entity.DocItem
		for i := 0; i < num_of_items; i++ {
			docItems = append(docItems, *createRandomDocItem(i + 1))
		}
		doc_res.AddDocItems = docItems
	}
	if editDefiner[4] {
		num_of_items := rand.Intn(len(doc.DocItems)/5) + 1
		var docItems []uint64
		var chosen []int
		for i := 0; i < num_of_items; i++ {
			t := rand.Intn(len(doc.DocItems))
			flag := 0
			for j := 0; j < len(chosen); j++ {
				if t == chosen[j] {
					flag = 1
					break
				}
			}
			if flag == 0 {
				docItems = append(docItems, doc.DocItems[t].ID)
				chosen = append(chosen, t)
			}
		}
		doc_res.RemoveDocItems = docItems
	}
	if editDefiner[5] {
		num_of_items := rand.Intn(len(doc.DocItems)/3) + 1
		var docItems []entity.DocItem
		var chosen []int
		for i := 0; i < num_of_items; i++ {
			t := rand.Intn(len(doc.DocItems))
			flag := 0
			for j := 0; j < len(doc_res.RemoveDocItems); j++ {
				if doc.DocItems[t].ID == doc_res.RemoveDocItems[j] {
					flag = 1
					break
				}
			}
			if flag == 0 {
				for j := 0; j < len(chosen); j++ {
					if t == chosen[j] {
						flag = 1
						break
					}
				}
				if flag == 0 {
					docItems = append(docItems, *createRandomEditDocItem(doc.DocItems[t]))
					chosen = append(chosen, t)
				}
			}
		}
		doc_res.EditDocItems = docItems
	}
	return doc_res
}

func createRandomEditDocItem(d entity.DocItem) *entity.DocItem {
	var docItem entity.DocItem = entity.DocItem{ID: d.ID, Moein: d.Moein, Tafsili: d.Tafsili, Bedehkar: d.Bedehkar, Bestankar: d.Bestankar, Desc: d.Desc, Curr: d.Curr, CurrPrice: d.CurrPrice, CurrRate: d.CurrRate}
	if rand.Intn(5) == 0 {
		docItem.Moein = moeins[rand.Intn(10)]
	}
	if docItem.Moein.TrackPossible {
		if rand.Intn(3) == 0 {
			docItem.Tafsili = tafsilis[rand.Intn(9)+1]
		}
	} else {
		docItem.Tafsili = tafsilis[0]
	}
	if docItem.Moein.CurrPossible {
		if rand.Intn(5) == 0 {
			idx := rand.Intn(5)
			docItem.Curr = currs[idx]
			docItem.CurrPrice = rand.Intn(1000) + 10
			docItem.CurrRate = currRates[idx]
			if rand.Intn(2) == 0 {
				docItem.Bedehkar = docItem.CurrRate * docItem.CurrPrice
				docItem.Bestankar = 0
			} else {
				docItem.Bestankar = docItem.CurrRate * docItem.CurrPrice
				docItem.Bedehkar = 0
			}
		}
	} else {
		if rand.Intn(4) == 0 {
			if rand.Intn(2) == 0 {
				docItem.Bedehkar = rand.Intn(100) * 1000
				docItem.Bestankar = 0
			} else {
				docItem.Bestankar = rand.Intn(100) * 1000
				docItem.Bedehkar = 0
			}
		}
	}
	return &docItem
}
func createRandomDocItem(num int) *entity.DocItem {
	var docItem entity.DocItem
	docItem.Desc = descItems[rand.Intn(50)]
	docItem.Num = num
	docItem.SaveDB = false
	docItem.Moein = moeins[rand.Intn(10)]
	if docItem.Moein.TrackPossible {
		docItem.Tafsili = tafsilis[rand.Intn(9)+1]
	} else {
		docItem.Tafsili = tafsilis[0]
		docItem.TafsiliID = tafsilis[0].ID
		//fmt.Println("X", docItem.Tafsili.CodeVal)
	}
	if docItem.Moein.CurrPossible {
		idx := rand.Intn(5)
		docItem.Curr = currs[idx]
		docItem.CurrPrice = rand.Intn(1000) + 10
		docItem.CurrRate = currRates[idx]
		if rand.Intn(2) == 0 {
			docItem.Bedehkar = docItem.CurrRate * docItem.CurrPrice
			docItem.Bestankar = 0
		} else {
			docItem.Bestankar = docItem.CurrRate * docItem.CurrPrice
			docItem.Bedehkar = 0
		}
	} else {
		if rand.Intn(2) == 0 {
			docItem.Bedehkar = rand.Intn(100) * 1000
			docItem.Bestankar = 0
		} else {
			docItem.Bestankar = rand.Intn(100) * 1000
			docItem.Bedehkar = 0
		}
	}
	return &docItem
}

func equalDoc(doc1 entity.Doc, doc2 entity.Doc) string {
	if doc1.AtfNum != doc2.AtfNum {
		return "atf"
	}
	if doc1.DailyNum != doc2.DailyNum {
		return "daily"
	}
	if doc1.Year != doc2.Year || doc1.Month != doc2.Month || doc1.Day != doc2.Day {
		return "date"
	}
	if doc1.Hour != doc2.Hour || doc1.Minute != doc2.Minute || doc1.Second != doc2.Second {
		return "date"
	}
	if doc1.Desc != doc2.Desc {
		return "desc"
	}
	if doc1.DocNum != doc2.DocNum {
		return "doc_num"
	}
	if doc1.DocType != doc2.DocType {
		return "doc type"
	}
	if doc1.EmitSystem != doc2.EmitSystem {
		return "emit"
	}
	if doc1.IsChanging != doc2.IsChanging {
		return "ischange"
	}
	if doc1.State != doc2.State {
		return "state"
	}
	if doc1.MinorNum != doc2.MinorNum {
		return "minor"
	}
	for i := range doc1.DocItems {
		if equalDocItem(doc1.DocItems[i], doc2.DocItems[i]) == "" {
			return "docitem" + "_" + equalDocItem(doc1.DocItems[i], doc2.DocItems[i])
		}
	}
	return ""
}

func equalDocItem(item1 entity.DocItem, item2 entity.DocItem) string {
	if item1.Bedehkar != item2.Bedehkar {
		return "bedeh"
	}
	if item1.Bestankar != item2.Bestankar {
		return "bestan"
	}
	if item1.Curr != item2.Curr {
		return "curr"
	}
	if item1.CurrPrice != item2.CurrPrice {
		return "currPr"
	}
	if item1.CurrRate != item2.CurrRate {
		return "currRa"
	}
	if item1.Desc != item2.Desc {
		return "desc"
	}
	if item1.Num != item2.Num {
		return "num"
	}
	if !equalCode(&item1.Moein, &item2.Moein) {
		return "moein"
	}
	if !equalCode(&item1.Tafsili, &item2.Tafsili) {
		return "tafsil"
	}
	return ""
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

func ClearTable(DocService *DocService) {
	// (*DocService).GetDB().Exec("DELETE FROM docs")
	// (*DocService).GetDB().Exec("ALTER SEQUENCE docs_id_seq RESTART WITH 1")
	// (*DocService).GetDB().Exec("DELETE FROM doc_items")
	// (*DocService).GetDB().Exec("ALTER SEQUENCE doc_items_id_seq RESTART WITH 1")
	// (*DocService).GetDB().Exec("DELETE FROM moeins")
	// (*DocService).GetDB().Exec("ALTER SEQUENCE moeins_id_seq RESTART WITH 1")
	// (*DocService).GetDB().Exec("DELETE FROM tafsilis")
	// (*DocService).GetDB().Exec("ALTER SEQUENCE tafsilis_id_seq RESTART WITH 1")
	// (*DocService).GetDB().Exec("DELETE FROM global_vars")
	// (*DocService).GetDB().Exec("ALTER SEQUENCE global_vars_id_seq RESTART WITH 1")

	(*DocService).GetDB().Migrator().DropTable(&entity.Doc{})
	(*DocService).GetDB().Migrator().DropTable(&entity.DocItem{})
	(*DocService).GetDB().Migrator().DropTable(&entity.Moein{})
	(*DocService).GetDB().Migrator().DropTable(&entity.Tafsili{})
	(*DocService).GetDB().Migrator().DropTable(&entity.GlobalVars{})

	// (*DocService).GetDB().Where("1 = 1").Delete(&entity.Doc{})
	// (*DocService).GetDB().Where("1 = 1").Delete(&entity.DocItem{})
	// (*DocService).GetDB().Where("1 = 1").Delete(&entity.GlobalVars{})
	// (*DocService).GetDB().Where("1 = 1").Delete(&entity.Moein{})
	// (*DocService).GetDB().Where("1 = 1").Delete(&entity.Tafsili{})

}
