package service

import (
	"AccountingDoc/Gin-Server/entity"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

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
	docs := []entity.Doc{}

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
			doc, err := insertRandomTestDoc(DocService)
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
		ClearTable(&DocService)
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

func TestSave(t *testing.T) {
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
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}

	var doc entity.Doc
	num_of_items := 10
	docItems := []entity.DocItem{}
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
	doc.Day = 29
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
	err = DocService.Save(doc)
	if err != nil {
		t.Error(err.Error())
	}
	//reqBody, err = json.Marshal(doc)
}

/*func FuzzSave(f *testing.F) {
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

	f.Fuzz(func(t *testing.T, num_of_items int, month int, day int, desc string, minor_num string) {
		var err error
		moeins, tafsilis, err = CreateMoeinTafsili(DocService)
		if err != nil {
			t.Fail()
		}

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
		var doc entity.Doc
		var docItems []entity.DocItem = make([]entity.DocItem, num_of_items)
		doc.DocItems = make([]entity.DocItem, num_of_items)
		t.Log(num_of_items, month, day, desc, minor_num)
		for i := 0; i < num_of_items; i++ {
			docItems = append(docItems, *createRandomDocItem(i + 1))
		}
		doc.DocItems = docItems
		doc.AtfNum = atf_num_global
		//atf_num_global += 1
		doc.DailyNum = daily_num_global
		//daily_num_global += 1
		doc.Year = 1401
		doc.Month = month
		doc.Day = day
		doc.Hour = rand.Intn(23) + 1
		doc.Minute = rand.Intn(59) + 1
		doc.Second = rand.Intn(59) + 1
		doc.Desc = desc
		doc.DocNum = 0
		doc.DocType = "عمومی"
		doc.EmitSystem = "سیستم حسابداری"
		doc.IsChanging = false
		doc.MinorNum = minor_num
		doc.State = "موقت"
		err = DocService.Save(doc)
		if err != nil {
			if month <= 12 && month > 0 && day > 0 && day < 32 && checkMinor(minor_num) {
				t.Fail()
			}
		}
	})
}*/

func checkMinor(str string) bool {
	for i := 0; i < len(str); i++ {
		fmt.Println(string(str[i]))
		_, err := strconv.Atoi(string(str[i]))
		if err != nil {
			return false
		}
	}
	return true
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
		ClearTable(&DocService)
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
		ClearTable(&DocService)
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
				//t.Log(actual.Message)
				t.Errorf("HTTP request status code error")
			}
			var actual entity.Doc
			if err := json.Unmarshal(body, &actual); err != nil {
				t.Errorf(err.Error())
			}
			if equalDoc(actual, docs[idx]) != "" {
				//t.Logf(equalDoc(actual, docs[idx]))
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
	r := 100

	editReqs := make([]int, n)

	rand.Seed(int64(n))
	docs, err := Init(DocService, n)
	if err != nil {
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}

	t.Log(len(docs))
	ch := make(chan struct{}, r)
	mu2 = make([]sync.RWMutex, n)

	for i := 0; i < r; i++ {
		go func() {
			idx := rand.Intn(n)

			err := DocService.CanEdit(uint64(idx + 1)) //SetupCanEditDoc(DocService, idx+1)
			if err == nil {
				change(&editReqs[idx], &docs[idx], idx)
				doc_edit := createRandomEditDoc(docs[idx])
				//time.Sleep(200 * time.Millisecond)
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
					if actual.Message != "date must be bigger than latest permanent doc" {
						t.Errorf(actual.Message)
					}
				} else {
					docs[idx].AtfNum = doc_edit.AtfNum
					docs[idx].DailyNum = doc_edit.DailyNum
					docs[idx].Year = doc_edit.Year
					docs[idx].Month = doc_edit.Month
					docs[idx].Day = doc_edit.Day
					docs[idx].Hour = doc_edit.Hour
					docs[idx].Minute = doc_edit.Minute
					docs[idx].Second = doc_edit.Second
					docs[idx].DocNum = doc_edit.DocNum
					docs[idx].MinorNum = doc_edit.MinorNum
					docs[idx].Desc = doc_edit.Desc
					docs[idx].State = doc_edit.State
					docs[idx].DocType = doc_edit.DocType
					docs[idx].EmitSystem = doc_edit.EmitSystem
					sort.Slice(doc_edit.RemoveDocItems, func(i2, j int) bool {
						return doc_edit.RemoveDocItems[i2] > doc_edit.RemoveDocItems[j]
					})
					for _, item := range doc_edit.RemoveDocItems {
						for j, item2 := range docs[idx].DocItems {
							if item == item2.ID {
								docs[idx].DocItems = append(docs[idx].DocItems[:j], docs[idx].DocItems[j:]...)
								break
							}
						}
					}
					for _, item := range doc_edit.EditDocItems {
						for j, item2 := range docs[idx].DocItems {
							if item.ID == item2.ID {
								docs[idx].DocItems[j] = item
								docs[idx].DocItems[j].Moein = item.Moein
								docs[idx].DocItems[j].Tafsili = item.Tafsili
								break
							}
						}
					}
					docs[idx].DocItems = append(docs[idx].DocItems, doc_edit.AddDocItems...)

					err := DocService.ChangeIsChange(uint64(idx + 1)) //SetupIsChangeDoc(DocService, idx+1)
					if err != nil {
						//t.Log(actual.Message)
						t.Errorf("HTTP After Edit request status code error")
					}
					afterChange(&editReqs[idx], idx)
				}
			} else {
				t.Log(err.Error())
				if docs[idx].State != "دائمی" {
					t.Log("Reject Not " + strconv.FormatInt(int64(idx+1), 10) + "\t" + strconv.FormatFloat(float64(time.Now().Nanosecond())/100000, 'e', 3, 64))
					notChange(&editReqs[idx], &docs[idx], idx)
				} else {
					if err.Error() != "doc is permanent" {
						t.Errorf("Reject permanent " + strconv.FormatInt(int64(idx+1), 10))
					}
				}
			}
			ch <- struct{}{}
		}()
	}

	for {
		idx := rand.Intn(n)

		err = DocService.CanEdit(uint64(idx + 1))
		if err == nil {
			change(&editReqs[idx], &docs[idx], idx)
			doc_edit := createRandomEditDoc(docs[idx])
			doc_edit.Month = 3
			doc_edit.AddDocItems = make([]entity.DocItem, 0)
			doc_edit.EditDocItems = make([]entity.DocItem, 0)
			doc_edit.RemoveDocItems = make([]uint64, 0)
			var doc_temp entity.Doc
			DocService.GetDB().Model(&entity.Doc{}).Preload("DocItems").Where("id = ?", idx+1).Find(&doc_temp)
			for _, item := range doc_temp.DocItems {
				doc_edit.RemoveDocItems = append(doc_edit.RemoveDocItems, item.ID)
			}
			//time.Sleep(1 * time.Second)
			_, w := SetupPutDoc(DocService, doc_edit)
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error() + "_last")
			}
			t.Log(len(doc_edit.AddDocItems), len(doc_edit.EditDocItems), len(doc_edit.RemoveDocItems), len(docs[idx].DocItems))
			if w.Code != http.StatusOK {
				var actual AnsReq
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Errorf(err.Error() + "_last")
				}
				if actual.Message != "it must have at least 1 doc item" {
					t.Errorf(actual.Message + "_last")
				}
			} else {
				t.Errorf("must not accept" + "_last " + strconv.FormatInt(int64(idx+1), 10))
			}
			err = DocService.ChangeIsChange(uint64(idx + 1)) //SetupIsChangeDoc(DocService, idx+1)
			if err != nil {
				t.Errorf("HTTP After Edit request status code error" + "_last")
			}
			afterChange(&editReqs[idx], idx)
			break
		} else {
			t.Log(err.Error())
			if docs[idx].State != "دائمی" {
				t.Log("Reject Not " + strconv.FormatInt(int64(idx+1), 10) + "\t" + strconv.FormatFloat(float64(time.Now().Nanosecond())/100000, 'e', 3, 64) + "_last")
				notChange(&editReqs[idx], &docs[idx], idx)
			} else {
				t.Log("Reject Permanent" + strconv.FormatInt(int64(idx+1), 10) + "_last")
			}
		}
	}

	for i := 0; i < r; i++ {
		<-ch
	}

	actual, err := DocService.FindAll()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		for i := 0; i < len(docs); i++ {
			if equalDoc(actual[i], docs[actual[i].AtfNum-1]) != "" && equalDoc(actual[i], docs[actual[i].AtfNum-1]) != "ischange" {
				t.Errorf(equalDoc(actual[i], docs[actual[i].AtfNum-1]))
			}
		}
	}

}

func change(x *int, y *entity.Doc, idx int) {
	mu2[idx].Lock()
	defer mu2[idx].Unlock()
	if *x != 0 {
		panic(y.ID)
		//t.Errorf("more than one changing doc with id %d number of %d", idx+1, editReqs[idx])
	}
	*x += 1
}

func afterChange(x *int, idx int) {
	mu2[idx].Lock()
	defer mu2[idx].Unlock()
	*x -= 1
}

func notChange(x *int, y *entity.Doc, idx int) {
	mu2[idx].Lock()
	defer mu2[idx].Unlock()
	if *x != 1 {
		panic("problem in changing doc with id " + strconv.FormatInt(int64(idx)+1, 10) + " with number of " + strconv.FormatInt(int64(*x), 10) + "------- " + strconv.FormatBool(y.IsChanging))
		//t.Errorf("problem in changing doc with id %d with number of %d", idx+1, editReqs[idx])
	}
}

func Test_ChangeState(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 10
	r := 100

	editReqs := make([]int, n)

	rand.Seed(int64(n))
	docs, err := Init(DocService, n)
	if err != nil {
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}

	t.Log(len(docs))
	ch := make(chan struct{}, r)
	mu2 = make([]sync.RWMutex, n)

	for i := 0; i < r; i++ {
		go func() {
			idx := rand.Intn(n)

			_, w := SetupPutChangeStateDoc(DocService, idx+1)
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error())
			}
			if w.Code != http.StatusOK {
				var actual AnsReq
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Errorf(err.Error())
				}
				if docs[idx].State != "دائمی" && actual.Message == "doc is permanent" {
					t.Errorf(actual.Message + " موقت")
				}
				if actual.Message == "sb else is changing doc" {
					notChange(&editReqs[idx], &docs[idx], idx)
				}
			} else {
				change(&editReqs[idx], &docs[idx], idx)
				docs[idx].State = "دائمی"
				afterChange(&editReqs[idx], idx)
			}
			ch <- struct{}{}
		}()
	}

	for i := 0; i < r; i++ {
		<-ch
	}

	actual, err := DocService.FindAll()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		for i := 0; i < len(docs); i++ {
			if equalDoc(actual[i], docs[actual[i].AtfNum-1]) != "" && equalDoc(actual[i], docs[actual[i].AtfNum-1]) != "ischange" {
				t.Errorf(equalDoc(actual[i], docs[actual[i].AtfNum-1]))
			}
		}
	}

}

func Test_DeleteDoc(t *testing.T) {
	db := NewDbConnectionTest()
	glb := &entity.GlobalVars{}
	glb.AtfNumGlobal = 1
	glb.TodayCount = 1
	db.Create(glb)

	DocService := New(db)
	defer DocService.CloseDB()
	defer ClearTable(&DocService)
	n := 100
	r := 1
	rand.Seed(int64(n))
	docs, err := Init(DocService, n)
	first_docs := docs
	if err != nil {
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}
	editReqs := make([]int, len(docs))
	ch := make(chan struct{}, r)
	mu2 = make([]sync.RWMutex, len(docs))

	for i := 0; i < r; i++ {
		go func() {
			idx := rand.Intn(len(docs))
			_, w := SetupDeleteDoc(DocService, first_docs[idx].AtfNum)
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error())
			}
			if w.Code != http.StatusOK {
				var actual AnsReq
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Errorf(err.Error())
				}
				if actual.Message != "record not found" {
					if docs[docs[idx].AtfNum-1].State != "دائمی" && actual.Message == "doc is permanent" {
						t.Errorf(actual.Message + " موقت")
					}
					if actual.Message == "sb else is changing doc" {
						notChange(&editReqs[docs[idx].AtfNum-1], &docs[idx], docs[idx].AtfNum-1)
					}
				}
			} else {
				change(&editReqs[docs[idx].AtfNum-1], &docs[idx], docs[idx].AtfNum-1)
				docs = append(docs[:idx], docs[idx+1:]...)
				t.Log(len(docs))
				afterChange(&editReqs[docs[idx].AtfNum-1], docs[idx].AtfNum-1)
			}
			ch <- struct{}{}
		}()
	}

	for i := 0; i < r; i++ {
		<-ch
	}

	actual, err := DocService.FindAll()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		t.Log(len(actual), len(docs))
		for i := 0; i < len(docs); i++ {
			flag := 0
			for j := 0; j < len(actual); j++ {
				if equalDoc(actual[j], docs[i]) == "" || equalDoc(actual[j], docs[i]) == "ischange" {
					flag = 1
					break
				}
			}
			if flag == 0 {
				t.Fail()
			}
		}
	}

}

func Test_GetMoeinTafsili(t *testing.T) {
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
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}
	DocService.GetDB().Save(&moeins)
	DocService.GetDB().Save(&tafsilis)

	req, w := SetupGetMoein(DocService)
	if req.Method != http.MethodGet {
		t.Errorf("HTTP request method error")
	}
	if w.Code != http.StatusOK {
		t.Errorf("HTTP request status code error")
	}
	req2, w2 := SetupGetTafsili(DocService)
	if req2.Method != http.MethodGet {
		t.Errorf("HTTP request method error")
	}
	if w2.Code != http.StatusOK {
		t.Errorf("HTTP request status code error")
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}

	actual := []entity.Moein{}
	if err := json.Unmarshal(body, &actual); err != nil {
		t.Errorf(err.Error())
	}

	for _, item := range moeins {
		for _, item2 := range actual {
			if item.CodeVal == item2.CodeVal {
				if item.CurrPossible != item2.CurrPossible || item.Name != item2.Name || item.TrackPossible != item2.TrackPossible {
					t.Errorf("moeins dont match")
				}
				break
			}
		}
	}

	body, err = ioutil.ReadAll(w2.Body)
	if err != nil {
		t.Errorf(err.Error())
	}

	actual2 := []entity.Tafsili{}
	if err := json.Unmarshal(body, &actual2); err != nil {
		t.Errorf(err.Error())
	}

	for _, item := range tafsilis {
		for _, item2 := range actual2 {
			if item.CodeVal == item2.CodeVal {
				if item.CurrPossible != item2.CurrPossible || item.Name != item2.Name || item.TrackPossible != item2.TrackPossible {
					t.Errorf("tafsilis dont match")
				}
				break
			}
		}
	}
}

func Test_ValidateDocItem(t *testing.T) {
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
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}
	DocService.GetDB().Save(&moeins)
	DocService.GetDB().Save(&tafsilis)

	r := 10
	for i := 0; i < r; i++ {
		go func() {
			docItem := createRandomDocItem(1)
			_, w := SetupValidateDocItem(DocService, *docItem)
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				t.Errorf(err.Error())
			}
			if w.Code != http.StatusOK {
				var actual AnsReq
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Errorf(err.Error())
				}
				//t.Log(*docItem)
				t.Errorf(actual.Message)
			}
		}()
	}

	docItem := createRandomDocItem(1)
	docItem.Desc = ""
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "docItem needs a description" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}

	//
	docItem = createRandomDocItem(1)
	docItem.Bedehkar = 0
	docItem.Bestankar = 0
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "docItem must have value in bestankar or bedehkar" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	docItem.Bedehkar = 100
	docItem.Bestankar = 1000
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "docItem must have value in one of bestankar and bedehkar fields" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	docItem.Moein.CodeVal = "8888p"
	err = validNegDocItem(DocService, *docItem)
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if item.TrackPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.Tafsili.CodeVal = "8888p"
	err = validNegDocItem(DocService, *docItem)
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if item.TrackPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.Tafsili.CodeVal = ""
	err = validNegDocItem(DocService, *docItem)
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if !item.TrackPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.Tafsili.CodeVal = tafsilis[1].CodeVal
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "moein cannot have any tafsilis" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if item.CurrPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.CurrPrice = 0
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "moein must have currency options" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if item.CurrPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.Curr = currs[1]
	docItem.CurrPrice = 3
	docItem.CurrRate = 7
	docItem.Bedehkar = 22
	docItem.Bestankar = 0
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "value in bedehkar doesnt match with currency" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if item.CurrPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.Curr = currs[1]
	docItem.CurrPrice = 3
	docItem.CurrRate = 7
	docItem.Bedehkar = 0
	docItem.Bestankar = 22
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "value in bestankar doesnt match with currency" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
	docItem = createRandomDocItem(1)
	for _, item := range moeins {
		if !item.CurrPossible {
			docItem.Moein = item
			break
		}
	}
	docItem.CurrRate = 5
	err = validNegDocItem(DocService, *docItem)
	if err != nil && err.Error() != "moein must not have currency options" {
		t.Errorf(err.Error())
	}
	if err == nil {
		t.Fail()
	}
	//
}

func Test_Numbering(t *testing.T) {
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
	docs, err := Init(DocService, n)
	if err != nil {
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}

	_, w := SetupNumbering(DocService)
	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Errorf(err.Error())
	}
	if w.Code != http.StatusOK {
		var actual AnsReq
		if err := json.Unmarshal(body, &actual); err != nil {
			t.Errorf(err.Error())
		}
	}

	DocService.GetDB().Model(&entity.Doc{}).Order("doc_num asc").Find(&docs)
	for i := 0; i < len(docs)-1; i++ {
		d1 := Date{docs[i].Year, docs[i].Month, docs[i].Day, docs[i].Hour, docs[i].Minute, docs[i].Second}
		d2 := Date{docs[i+1].Year, docs[i+1].Month, docs[i+1].Day, docs[i+1].Hour, docs[i+1].Minute, docs[i+1].Second}
		if compare(&d1, &d2) > 0 {
			t.Fail()
		}
	}

}

func Test_FilterByMinorNum(t *testing.T) {
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
	docs, err := Init(DocService, n)
	if err != nil {
		ClearTable(&DocService)
		t.Errorf(err.Error())
	}

	m := make(map[string]int)
	for _, item := range docs {
		x, ok := m[item.MinorNum]
		if !ok {
			m[item.MinorNum] = 1
		} else {
			m[item.MinorNum] = x + 1
		}
	}

	for i := 0; i < 6; i++ {
		minorNum := minorNums[i]
		_, w := SetupFilterMinor(DocService, minorNum)
		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf(err.Error())
		}
		if i == 0 {
			if w.Code != 500 {
				t.Fail()
			}
		} else {
			if w.Code == 500 {
				t.Fail()
			} else {
				var actual []entity.Doc
				if err := json.Unmarshal(body, &actual); err != nil {
					t.Fail()
				}
				if len(actual) != m[minorNums[i]] {
					t.Fail()
				}
			}
		}
	}
}

func validNegDocItem(DocService DocService, docItem entity.DocItem) error {
	_, w := SetupValidateDocItem(DocService, docItem)
	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		return err
	}
	if w.Code == http.StatusOK {
		return nil
	} else {
		var actual AnsReq
		if err := json.Unmarshal(body, &actual); err != nil {
			return err
		}
		if actual.Message != "" {
			return errors.New(actual.Message)
		}
	}
	return nil
}

func SetupGetDocs(DocService DocService) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
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

func SetupFilterMinor(DocService DocService, idx string) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/filter/minor_num/" + idx
	r.GET("/docs/filter/minor_num/", func(c *gin.Context) {
		c.JSON(500, gin.H{
			"message": "minor num field must not be empty",
		})
	})
	r.GET("/docs/filter/minor_num/:minor", func(c *gin.Context) {
		minor_num := c.Param("minor")
		res, err := DocService.FilterByMinorNum(minor_num)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(200, res)
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

func SetupNumbering(DocService DocService) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/numbering"
	r.GET("/docs/numbering", func(c *gin.Context) {
		err := DocService.Numbering()
		if err == nil {
			c.JSON(200, gin.H{
				"message": "Successfully Numbered",
			})
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

func SetupGetMoein(DocService DocService) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/moeins"
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupGetTafsili(DocService DocService) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/tafsilis"
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
	gin.SetMode(gin.ReleaseMode)
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
func SetupDeleteDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/" + strconv.FormatInt(int64(idx), 10)
	r.DELETE("/docs/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err := DocService.DeleteByID(id)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Successfully Delete",
				})
			} else {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			}
		}
	})

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupCanEditDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
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

func SetupPutChangeStateDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	url := "/docs/change_state/" + strconv.FormatInt(int64(idx), 10)
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
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupGetDoc(DocService DocService, idx int) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
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
	gin.SetMode(gin.ReleaseMode)
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

func SetupValidateDocItem(DocService DocService, docItem entity.DocItem) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/docs/validate_doc_item", func(c *gin.Context) {
		var docItem entity.DocItem
		err := c.ShouldBindJSON(&docItem)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		} else {
			err = DocService.ValidateDocItem(docItem)
			if err != nil {
				c.JSON(500, gin.H{
					"message": err.Error(),
				})
			} else {
				c.JSON(200, entity.Codes{Moein: docItem.Moein, Tafsili: docItem.Tafsili})
			}
		}
	})
	reqBody, err := json.Marshal(docItem)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/docs/validate_doc_item", bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return req, w
}

func SetupPutDoc(DocService DocService, doc entity.AddRemoveDocItem) (*http.Request, *httptest.ResponseRecorder) {
	gin.SetMode(gin.ReleaseMode)
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

func insertRandomTestDoc(DocService DocService) (entity.Doc, error) {
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
