package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gilwo/Sh0r7/store"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// func TestCreateShortDataImproved1__(t *testing.T) {
// 	var status int
// 	var body interface{}
// 	c, _ := gin.CreateTestContext(httptest.NewRecorder())
// 	c.JSON = func(stat int, object interface{}) {
// 		status = stat
// 		body = object
// 	}
// 	CreateShortDataImproved1(c)
// 	assert.Equal(t, 4, 4)
// }

type maybeJsonObj struct {
	asObj    map[string]interface{}
	asString string
}

func newFromString(m string) *maybeJsonObj {
	// fmt.Printf("!!!!!!!!!!! %s\n", m)
	r := &maybeJsonObj{asObj: map[string]interface{}{}}
	err := json.Unmarshal([]byte(m), &r.asObj)
	if err != nil {
		r.asString = m
		return r
	}
	data, err := json.Marshal(r.asObj)
	if err != nil {
		panic(fmt.Errorf("convert map to json string failed: %s", err))
	}
	r.asString = string(data)
	return r
}

func newFromMap(m map[string]interface{}) *maybeJsonObj {
	data, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Errorf("convert map to json string failed: %s", err))
	}
	r := &maybeJsonObj{asObj: map[string]interface{}{}, asString: string(data)}
	err = json.Unmarshal([]byte(r.asString), &r.asObj)
	if err != nil {
		panic(fmt.Errorf("convert json string to map failed: %s", err))
	}
	return r
}
func (jo *maybeJsonObj) String() string {
	return jo.asString
}
func (jo *maybeJsonObj) Map() map[string]interface{} {
	return jo.asObj
}
func (jo *maybeJsonObj) Comp(other *maybeJsonObj) bool {
	fmt.Printf("we: %s\n", jo.asString)
	fmt.Printf("other: %s\n", other.asString)
	return jo.asString == other.asString
}

func test_init(t *testing.T) *gin.Engine {
	store.StoreCtx = store.NewStoreLocal()
	assert.NoError(t, store.StoreCtx.InitializeStore())
	return GinInit()
}

func TestCreateShortData(t *testing.T) {
	PAYLOAD := newFromString(`{"url": "yahoo.com"}`)
	ge := test_init(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/create-short-data", bytes.NewBufferString(PAYLOAD.String()))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	fmt.Printf("%v\n", w.Body)
	res := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "delete")
	assert.Contains(t, res, "modify")
	assert.Contains(t, res, "short")
}

func TestDummyFail(t *testing.T) {
	ge := test_init(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dummy/data", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	// HandleGetOriginData(c)
	ge.ServeHTTP(w, r)
	assert.Equal(t, 400, w.Code) // or what value you need it to be

	fmt.Printf("%v", w.Body)
	res := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "error")
	assert.Contains(t, res["error"], "entry not exist for dummy")
}

// func TestCreateShortDataImproved1(t *testing.T) {
// 	router := setupRouter()

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/ping", nil)
// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)
// 	assert.Equal(t, "pong", w.Body.String())
// }

func TestFullFlowUrl(t *testing.T) {
	var (
		PAYLOAD        = newFromString(`{"url": "google.com"}`)
		RESPPAYLOAD    = newFromString("google.com")
		MODPAYLOAD     = newFromString(`{"url": "cnn.com"}`)
		RESPMODPAYLOAD = newFromString("cnn.com")
	)
	testFullFlow(t, "/create-short-url", PAYLOAD, RESPPAYLOAD, MODPAYLOAD, RESPMODPAYLOAD)
}

func TestFullFlowData(t *testing.T) {
	var (
		PAYLOAD        = newFromString(`{"url": "google.com"}`)
		RESPPAYLOAD    = newFromString(`{"url": "google.com"}`)
		MODPAYLOAD     = newFromString(`{"url": "cnn.com"}`)
		RESPMODPAYLOAD = newFromString(`{"url": "cnn.com"}`)
	)
	testFullFlow(t, "/create-short-data", PAYLOAD, RESPPAYLOAD, MODPAYLOAD, RESPMODPAYLOAD)
}

func testFullFlow(t *testing.T, api string, req1, resp1, req2, resp2 *maybeJsonObj) {
	ge := test_init(t)

	// create
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", api, bytes.NewBufferString(req1.String()))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	fmt.Printf("%v\n", w.Body)
	res := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "short")
	assert.Contains(t, res, "modify")
	assert.Contains(t, res, "delete")
	short := res["short"].(string)
	modify := res["modify"].(string)
	delete := res["delete"].(string)

	fmt.Printf("checking short %s info\n", short)

	// retrieve info
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/info", short), nil)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	res = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "created")
	assert.Contains(t, res, "data")

	// retrieve data
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/data", short), nil)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	// res = map[string]interface{}{}
	// assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	// assert.Contains(t, res, "data")
	act := newFromString(w.Body.String())
	assert.True(t, act.Comp(resp1))

	// patch data
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PATCH", fmt.Sprintf("/%s", modify), bytes.NewBufferString(req2.String()))
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	// fmt.Printf("%v\n", w.Body)
	// res = map[string]interface{}{}
	// assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	// assert.Contains(t, res, "data")
	// assert.Equal(t, res["data"], PAYLOAD)

	// retrieve data - check data is the mod payload
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/data", short), nil)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	// res = map[string]interface{}{}
	// assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	// assert.Contains(t, res, "data")
	act = newFromString(w.Body.String())
	assert.True(t, act.Comp(resp2))

	// retrieve info - check old data exists
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/info", short), nil)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	res = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "data_0")
	assert.Contains(t, res, "changed")
	assert.Equal(t, res["changed"], "1")

	// delete
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", fmt.Sprintf("/%s", delete), nil)
	// r.Header.Add("token", token)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	// res = map[string]interface{}{}
	// assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	// assert.Contains(t, res, "data_0")
	// assert.Contains(t, res, "changed")
	// assert.Equal(t, res["changed"], "1")

	// verify short not exits
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/data", short), nil)
	// r.Header.Add("token", token)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 400, w.Code) // or what value you need it to be
	fmt.Printf("%v", w.Body)
	res = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "error")
	assert.Contains(t, res["error"], fmt.Sprintf("entry not exist for %s", short))

}
