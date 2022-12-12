package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	"github.com/gin-gonic/gin"
	"github.com/karrick/tparse"
	"github.com/pkg/errors"
)

func _spawnErrWithCode(c *gin.Context, code int, err error) {
	fmt.Printf("err : %v\n", err)
	c.JSON(code, gin.H{"error": err.Error()})
}
func _spawnErr(c *gin.Context, err error) {
	_spawnErrWithCode(c, http.StatusBadRequest, err)
}
func _spawnErrCond(c *gin.Context, err error, cond bool) {
	if cond {
		_spawnErrWithCode(c, http.StatusBadRequest, err)
	}
}

var (
	runningCount = 0
)

func handleCreateShortModDelete(data string, isUrl bool, expiration time.Duration) (map[string]interface{}, error) {
	var err error
	res := map[string]interface{}{}
	shorts := map[string]string{
		"":  shortener.GenerateShortDataWithStore(data, store.StoreCtx),
		"d": shortener.GenerateShortDataWithStore(data+"delete", store.StoreCtx),
		"p": shortener.GenerateShortDataWithStore(data+"private", store.StoreCtx),
	}
	for _, e := range shorts {
		if e == "" {
			return nil, fmt.Errorf("there was a problem creating a short")
		}
	}
	mapping := map[string]string{
		"":  data,
		"d": shorts[""],
		"p": shorts[""],
	}
	if isUrl {
		shorts["url"] = shorts[""]
		mapping["url"] = shorts[""]
	}
	for k, e := range shorts {
		err = store.StoreCtx.SaveDataMapping([]byte(mapping[k]), e+k, expiration)
		if err != nil {
			break
		}
	}
	shorts["s"] = shorts[""]
	if err == nil {
		for k, e := range shorts {
			if k == "" {
				continue
			}
			err = store.StoreCtx.SetMetaDataMapping(shorts[""], k, e)
			if err != nil {
				break
			}
		}

	}
	if err != nil {
		for k, e := range shorts {
			_ = store.StoreCtx.RemoveDataMapping(e + k)
		}
		return nil, fmt.Errorf("there was a problem storing a short")
	}

	res["short"] = shorts[""]
	res["private"] = shorts["p"]
	res["delete"] = shorts["d"]
	return res, nil
}
func HandleCreateShortData(c *gin.Context) {
	if !checkToken(c) {
		return
	}
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	res, err := handleCreateShortModDelete(string(d), false, getExpiration(c))
	if err != nil {
		_spawnErr(c, err)
		return
	}
	fmt.Printf("res: %#v\n", res)
	c.JSON(200, res)
}

func HandleUploadFile(c *gin.Context) {
	adminKey := c.Query("adKey")
	if adminKey == "" {
		adminKey = c.Request.Header.Get("adKey")
	}
	adTok := shortener.GenerateTokenTweaked(adminKey, 0, 32, 0)
	if adTok == "" {
		adTok = c.Query("adTok")
		if adTok == "" {
			adTok = c.Request.Header.Get("adTok")
		}
	}
	if adTok == "" || !store.StoreCtx.CheckExistShortDataMapping(adTok) {
		_spawnErrWithCode(c, http.StatusForbidden, fmt.Errorf("operation not allowed"))
		fmt.Printf("invalid adTok (%s)\n", adTok)
		return
	}

	name := c.Query("name")
	if name == "" {
		_spawnErr(c, fmt.Errorf("invalid empty name"))
		fmt.Printf("name file is empty\n")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		_spawnErr(c, err)
		fmt.Printf("err in form file, %s\n", err)
		return
	}
	if file.Size > 1<<20 {
		err = fmt.Errorf("size too big")
		_spawnErr(c, err)
		fmt.Printf("err in file upload, %s\n", err)
		return
	}
	src, err := file.Open()
	if err != nil {
		_spawnErr(c, err)
		fmt.Printf("err in file open, %s\n", err)
		return
	}
	defer src.Close()

	buf := bytes.NewBuffer(func() []byte { return []byte{} }())

	_, err = io.Copy(buf, src)
	if err != nil {
		_spawnErr(c, err)
		fmt.Printf("err in file upload, %s\n", err)
		return
	}

	store.StoreCtx.RemoveDataMapping(name)
	store.StoreCtx.SaveDataMapping(buf.Bytes(), name, -1)
}

func HandleCreateShortUrl(c *gin.Context) {
	if !checkToken(c) {
		return
	}
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	mapping := map[string]string{}
	err = json.Unmarshal(d, &mapping)
	if err != nil {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		fmt.Printf("err in json unmarshal, %s\n", err)
		return
	}
	if url, ok := mapping["url"]; !ok {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		fmt.Printf("json field url is missing, %#v\n", mapping)
		return
	} else {
		res, err := handleCreateShortModDelete(url, true, getExpiration(c))
		if err != nil {
			_spawnErr(c, err)
			return
		}
		fmt.Printf("res: %#v\n", res)
		c.JSON(200, res)
	}
}

func HandleUpdateShort(c *gin.Context) {
	short := c.Param("short")
	d, err := c.GetRawData()
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErr(c, msg)
		return
	}

	if updateUrl(c, d) {
		return
	}
	if updateData(c, d) {
		return
	}

	_spawnErr(c, fmt.Errorf("short %s not found", short))
}
func updateUrl(c *gin.Context, d []byte) bool {
	short := c.Param("short")

	mod, err := store.StoreCtx.LoadDataMapping(short + "p")
	if err != nil {
		fmt.Println(err)
		return false
	}
	data, err := store.StoreCtx.LoadDataMapping(string(mod) + "url")
	if err != nil {
		fmt.Println(err)
		return false
	}

	mapping := map[string]string{}
	err = json.Unmarshal(d, &mapping)
	if err != nil {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		fmt.Printf("err in json unmarshal, %s\n", err)
		return false
	}
	url, ok := mapping["url"]
	if !ok {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		fmt.Printf("json field url is missing, %#v\n", mapping)
		return false
	}

	err = store.StoreCtx.UpdateDataMapping([]byte(url), string(data))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
func updateData(c *gin.Context, d []byte) bool {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + "p")
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = store.StoreCtx.UpdateDataMapping(d, string(dataKey))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
func DeleteShortData(c *gin.Context) {
	handleRemove(c, true)
}
func tryDelete(c *gin.Context) bool {
	return handleRemove(c, false)
}
func handleRemove(c *gin.Context, withResponse bool) bool {
	short := c.Param("short")
	removeKey := short + "d"
	dataKey, err := store.StoreCtx.LoadDataMapping(removeKey)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	privateKey, err := store.StoreCtx.GetMetaDataMapping(string(dataKey), "p")
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	privateKey += "p"
	if err := store.StoreCtx.RemoveDataMapping(string(dataKey)); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	if err := store.StoreCtx.RemoveDataMapping(privateKey); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	if err := store.StoreCtx.RemoveDataMapping(removeKey); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	store.StoreCtx.RemoveDataMapping(string(dataKey) + "url")
	if withResponse {
		c.Status(200)
	}
	return true
}

func HandleGetShortDataInfo(c *gin.Context) {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + "p")
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	log.Printf("data key: %s, short %s\n", string(dataKey), short)
	data, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey))
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	c.JSON(200, data)
}
func HandleGetOriginData(c *gin.Context) {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + "p")
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	data, err := store.StoreCtx.LoadDataMapping(string(dataKey))
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}

	c.String(200, "%s", data)
}

func HandleShort(c *gin.Context) {
	short := c.Param("short")
	fmt.Printf("trying url for <%s>\n", short)
	if tryUrl(c) {
		return
	}
	fmt.Printf("trying data for <%s>\n", short)
	if getData(c) {
		return
	}
	fmt.Printf("trying private data for <%s>\n", short)
	if getDataPrivate(c) {
		return
	}
	fmt.Printf("trying delete for <%s>\n", short)
	if tryDelete(c) {
		return
	}

	_spawnErr(c, fmt.Errorf("short %s not found", short))
}
func getData(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s\n", msg, err)
		return false
	}
	c.String(200, "%s", data)
	return true
}
func getDataPrivate(c *gin.Context) bool {
	short := c.Param("short")
	privateKey, err := store.StoreCtx.LoadDataMapping(short + "p")
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		return false
	}
	data, err := store.StoreCtx.LoadDataMapping(string(privateKey))
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("%s, err: %s", msg, err)
		return false
	}
	c.String(200, "%s", data)
	return true
}
func tryUrl(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short + "url")
	if err != nil {
		fmt.Println(err)
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("load url: %s, err: %s", msg, err)
		return false
	}

	data, err = store.StoreCtx.LoadDataMapping(string(data))
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		fmt.Printf("load data: %s, err: %s", msg, err)
		return false
	}
	url := string(data)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	c.Redirect(302, url)
	return true
}

func checkToken(c *gin.Context) bool {
	token := c.Request.Header.Get("TID")
	log.Printf("token: <%s> (%d)\n", token, len(token))
	info, err := store.StoreCtx.LoadDataMappingInfo(token)
	if err != nil {
		log.Printf("failed to get info for token: <%s>\n", token)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}

	v, ok := info["created"]
	if !ok {
		log.Printf("failed to get created time value for on token: <%s>\n", token)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}
	when := v.(string)
	// if time.Parse(when)
	t, err := time.Parse(time.RFC3339, when)
	if err != nil {
		log.Printf("failed to get parsed created time (%s) value for on token: <%s>\n", when, token)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}
	fmt.Printf("created time for key: <%s> : before parse [%s], after parse [%s]\n", token, when, t)
	v, ok = info["ttl"]
	if !ok {
		log.Printf("failed to get ttl value for on token: <%s>\n", token)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}
	ttl, err := time.ParseDuration(v.(string))
	if err != nil {
		log.Printf("failed to get parsed duration time (%s) value for on token: <%s>\n", v, token)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}

	if time.Since(t) > ttl {
		log.Printf("token (%s) expired : created <%s> ttl <%s>\n", token, t, ttl)
		_spawnErrWithCode(c, http.StatusUnauthorized, errors.New("invalid token"))
		return false
	}
	return true
}

func getExpiration(c *gin.Context) time.Duration {
	expiration := c.Request.Header.Get("exp")
	if expiration == "n" {
		return -1
	}
	t1, err := tparse.AddDuration(time.Time{}, expiration)
	if err != nil {
		log.Printf("expiration [%s] parsing failed: %s\n", expiration, err)
		return 0
	}
	return t1.Sub(time.Time{})
}
