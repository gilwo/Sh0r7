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

const (
	PAYLOAD = `
{
	"long_url2": "something something something",
	"long_url": "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html",
	"user_id": "e0dba740-fc4b-4977-872c-d360239e6b10",
	"sub": {
		"a": 1,
		"b": "2"
	}
}
`
	MODPAYLOAD = "data changed"
)

func test_init(t *testing.T) *gin.Engine {
	store.StoreCtx = store.NewStoreLocal()
	assert.NoError(t, store.StoreCtx.InitializeStore())
	return GinInit()
}

func TestCreateShortDataImproved1(t *testing.T) {
	ge := test_init(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/create-short-data", bytes.NewBufferString(PAYLOAD))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	fmt.Printf("%v\n", w.Body)
	res := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "shortData")
	assert.Contains(t, res, "token")
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

func TestFullFlow(t *testing.T) {
	ge := test_init(t)

	// create
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/create-short-data", bytes.NewBufferString(PAYLOAD))
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code) // or what value you need it to be

	fmt.Printf("%v\n", w.Body)
	res := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "shortData")
	assert.Contains(t, res, "token")
	short := res["shortData"].(string)
	token := res["token"].(string)

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
	assert.Contains(t, res, "token")
	assert.Contains(t, res["token"], token)

	// retrieve data
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/%s/data", short), nil)
	fmt.Printf("****\n%#v\n****\n", r)
	c, _ = gin.CreateTestContext(w)
	c.Request = r
	ge.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	fmt.Printf("%v\n", w.Body)
	res = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "data")
	assert.Equal(t, res["data"], PAYLOAD)

	// patch data
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PATCH", fmt.Sprintf("/%s", short), bytes.NewBufferString(MODPAYLOAD))
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
	res = map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Contains(t, res, "data")
	assert.Equal(t, res["data"], MODPAYLOAD)

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
	r = httptest.NewRequest("DELETE", fmt.Sprintf("/%s", short), nil)
	r.Header.Add("token", token)
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
	r.Header.Add("token", token)
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
