package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	"github.com/gilwo/Sh0r7/webapp/common"
	"github.com/gin-gonic/gin"
	"github.com/karrick/tparse"
	"github.com/pkg/errors"
)

func _spawnErrWithCode(c *gin.Context, code int, err error) {
	log.Printf("err : %v\n", err)
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

const (
	LengthMinShortFree = 5
	LengthMaxShortFree = 9
	LengthMinPrivate   = 10
	LengthMaxPrivate   = 15
	LengthMinDelete    = 10
	LengthMaxDelete    = 15
)

func handleCreateShortModDelete(data string, isUrl bool, expiration time.Duration) (map[string]string, error) {
	var err error
	shorts := map[string]string{
		store.SuffixPublic: shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixPublic, -1, 0, LengthMinShortFree, LengthMaxShortFree, store.StoreCtx),
		store.SuffixRemove: shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixRemove, -1, 0, LengthMinDelete, LengthMaxDelete, store.StoreCtx),
		store.SuffixPrivate: shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixRemove, -1, 0, LengthMinPrivate, LengthMaxPrivate, store.StoreCtx),
	}
	for k, e := range shorts {
		if e == "" {
			log.Printf("problem with crating short for <%s>\n", k)
			return nil, errors.Errorf("there was a problem creating a short")
		}
	}
	mapping := map[string]string{
		store.FieldPublic:  data,
		store.FieldRemove:  shorts[store.SuffixRemove],
		store.FieldPrivate: shorts[store.SuffixPrivate],
	}
	if isUrl {
		shorts[store.SuffixURL] = shorts[store.SuffixPublic]
		mapping[store.FieldURL] = shorts[store.SuffixPublic]
	}
	for k, e := range shorts {
		if k == store.SuffixPublic {
			err = store.StoreCtx.SaveDataMapping([]byte(data), e+k, expiration)
		} else {
			err = store.StoreCtx.SaveDataMapping([]byte(shorts[store.SuffixPublic]), e+k, expiration)
		}
		if err != nil {
			log.Printf("save data mapping <%s> failed for <%s>\n", e, k)
			break
		}
	}
	if err == nil {
		for k, e := range mapping {
			if k == store.FieldPublic {
				e = shorts[store.SuffixPublic]
			}
			err = store.StoreCtx.SetMetaDataMapping(shorts[store.SuffixPublic]+store.SuffixPublic, k, e)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		for k, e := range shorts {
			errRem := store.StoreCtx.RemoveDataMapping(e + k)
			if errRem != nil {
				log.Printf("problem remving key <%s>, err: %s\n", e+k, err.Error())
			}
		}
		return nil, errors.Errorf("there was a problem storing a short, err: %s", err)
	}


	return map[string]string{
		store.FieldPublic:  shorts[store.SuffixPublic],
		store.FieldPrivate: shorts[store.SuffixPrivate],
		store.FieldRemove:  shorts[store.SuffixRemove],
	}, nil
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
	handleCreateHeaders(c, res)
	log.Printf("res: %#v\n", res)
	c.JSON(200, res)
}

func HandleUploadFile(c *gin.Context) {
	adminKey := c.Query(common.FAdminKey)
	if adminKey == "" {
		adminKey = c.Request.Header.Get(common.FAdminKey)
	}
	adTok := shortener.GenerateTokenTweaked(adminKey, 0, 32, 0)
	if adTok == "" {
		adTok = c.Query(common.FAdminToken)
		if adTok == "" {
			adTok = c.Request.Header.Get(common.FAdminToken)
		}
	}
	if adTok == "" || !store.StoreCtx.CheckExistShortDataMapping(adTok) {
		_spawnErrWithCode(c, http.StatusForbidden, errors.Errorf("operation not allowed"))
		log.Printf("invalid admin token (%s)\n", adTok)
		return
	}

	name := c.Query(common.FFileName)
	if name == "" {
		_spawnErr(c, errors.Errorf("invalid empty name"))
		log.Printf("name file is empty\n")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		_spawnErr(c, err)
		log.Printf("err in form file, %s\n", err)
		return
	}
	if file.Size > 1<<20 { // 1MB
		err = errors.Errorf("size too big")
		_spawnErr(c, err)
		log.Printf("err in file upload, %s\n", err)
		return
	}
	src, err := file.Open()
	if err != nil {
		_spawnErr(c, err)
		log.Printf("err in file open, %s\n", err)
		return
	}
	defer src.Close()

	buf := bytes.NewBuffer(func() []byte { return []byte{} }())

	_, err = io.Copy(buf, src)
	if err != nil {
		_spawnErr(c, err)
		log.Printf("err in file upload, %s\n", err)
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
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("err in json unmarshal, %s\n", err)
		return
	}
	if url, ok := mapping["url"]; !ok {
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("json field url is missing, %#v\n", mapping)
		return
	} else {
		res, err := handleCreateShortModDelete(url, true, getExpiration(c))
		if err != nil {
			_spawnErr(c, err)
			return
		}
		handleCreateHeaders(c, res)
		log.Printf("res: %#v\n", res)
		c.JSON(200, res)
	}
}

func HandleUpdateShort(c *gin.Context) {
	short := c.Param("short")
	d, err := c.GetRawData()
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}

	if updateUrl(c, d) {
		return
	}
	if updateData(c, d) {
		return
	}

	_spawnErr(c, errors.Errorf("short %s not found", short))
}
func updateUrl(c *gin.Context, d []byte) bool {
	short := c.Param("short")

	mod, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPrivate)
	if err != nil {
		log.Println(err)
		return false
	}
	data, err := store.StoreCtx.LoadDataMapping(string(mod) + store.SuffixURL)
	if err != nil {
		log.Println(err)
		return false
	}

	mapping := map[string]string{}
	err = json.Unmarshal(d, &mapping)
	if err != nil {
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("err in json unmarshal, %s\n", err)
		return false
	}
	url, ok := mapping["url"]
	if !ok {
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("json field url is missing, %#v\n", mapping)
		return false
	}

	err = store.StoreCtx.UpdateDataMapping([]byte(url), string(data))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
func updateData(c *gin.Context, d []byte) bool {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPrivate)
	if err != nil {
		log.Println(err)
		return false
	}

	err = store.StoreCtx.UpdateDataMapping(d, string(dataKey))
	if err != nil {
		log.Println(err)
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
	removeKey := short + store.SuffixRemove
	dataKey, err := store.StoreCtx.LoadDataMapping(removeKey)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	privateKey, err := store.StoreCtx.GetMetaDataMapping(string(dataKey)+store.SuffixPublic, store.FieldPrivate)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	privateKey += store.SuffixPrivate
	if err := store.StoreCtx.RemoveDataMapping(string(dataKey) + store.SuffixPublic); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	if err := store.StoreCtx.RemoveDataMapping(privateKey); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	if err := store.StoreCtx.RemoveDataMapping(removeKey); err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErrCond(c, msg, withResponse)
		return false
	}
	store.StoreCtx.RemoveDataMapping(string(dataKey) + store.SuffixURL)
	if withResponse {
		c.Status(200)
	}
	return true
}

func HandleGetShortDataInfo(c *gin.Context) {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPrivate)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	log.Printf("data key: %s, short %s\n", string(dataKey), short)

	storedPrvPassTok, e1 := store.StoreCtx.GetMetaDataMapping(string(dataKey)+store.SuffixPublic, store.FieldPrvPassTok)
	if e1 == nil && storedPrvPassTok != "" {
		storedPrvPassSalt, e2 := store.StoreCtx.GetMetaDataMapping(string(dataKey)+store.SuffixPublic, store.FieldPrvPassSalt)
		if e2 != nil {
			msg := errors.Errorf("short <%s> is locked but missing salt", dataKey)
			log.Printf("%s, e2: %s\n", msg, e2.Error())
			_spawnErr(c, msg)
			return
		}
		recvPassToken := c.Request.Header.Get(common.FPrvPassToken)
		if recvPassToken != storedPrvPassTok {
			msg := errors.Errorf("access denied to short <%s>", short)
			log.Printf("%s (pass tok),  recv <%s>, salt <%s>, stored <%s>\n",
				msg, recvPassToken, storedPrvPassSalt, storedPrvPassTok)
			_spawnErrWithCode(c, http.StatusForbidden, msg)
			return
		}
	}

	data, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	c.JSON(200, data)
}
func HandleGetOriginData(c *gin.Context) {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPrivate)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}
	data, err := store.StoreCtx.LoadDataMapping(string(dataKey) + store.SuffixPublic)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		_spawnErr(c, msg)
		return
	}

	c.String(200, "%s", data)
}

func HandleShort(c *gin.Context) {
	short := c.Param("short")
	log.Printf("trying url for <%s>\n", short)
	if tryUrl(c) {
		return
	}
	log.Printf("trying data for <%s>\n", short)
	if getData(c) {
		return
	}
	log.Printf("trying private data for <%s>\n", short)
	if getDataPrivate(c) {
		return
	}
	log.Printf("trying delete for <%s>\n", short)
	if tryDelete(c) {
		return
	}

	_spawnErr(c, errors.Errorf("short %s not found", short))
}
func getData(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPublic)
	if err != nil {
		// if short fail here, then try to get the data for the full path
		//  (some elements are stored this way using the storage provider)
		data, err = store.StoreCtx.LoadDataMapping(c.Request.URL.Path)
		if err == nil {
			c.String(200, string(data))
			return true
		}
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		return false
	}
	c.String(200, "%s", data)
	return true
}
func getDataPrivate(c *gin.Context) bool {
	short := c.Param("short")
	dataKey, err := store.StoreCtx.LoadDataMapping(short + store.SuffixPrivate)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		return false
	}
	data, err := store.StoreCtx.LoadDataMapping(string(dataKey) + store.SuffixPublic)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		return false
	}
	c.String(200, "%s", data)
	return true
}
func tryUrl(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short + store.SuffixURL)
	if err != nil {
		log.Println(err)
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("load url: %s, err: %s\n", msg, err)
		return false
	}

	data, err = store.StoreCtx.LoadDataMapping(string(data) + store.SuffixPublic)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("load data: %s, err: %s\n", msg, err)
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
	token := c.Request.Header.Get(common.FTokenID)
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
	log.Printf("created time for key: <%s> : before parse [%s], after parse [%s]\n", token, when, t)
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
	expiration := c.Request.Header.Get(common.FExpiration)
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

func handleCreateHeaders(c *gin.Context, res map[string]string) {
	var err error
	if desc := c.Request.Header.Get(common.FShortDesc); desc != "" {
		store.StoreCtx.SetMetaDataMapping(res[store.FieldPublic]+store.SuffixPublic, store.FieldDesc, desc)
	}
	if prvPassTok := c.Request.Header.Get(common.FPrvPassToken); prvPassTok != "" {
		token := c.Request.Header.Get(common.FTokenID)
		err = store.StoreCtx.SetMetaDataMapping(res[store.FieldPublic]+store.SuffixPublic, store.FieldPrvPassSalt, token)
		if err != nil {
			log.Printf("failed keeping prv pass token on short <%s> metadata\n", res[store.FieldPublic])
		}
		err = store.StoreCtx.SetMetaDataMapping(res[store.FieldPublic]+store.SuffixPublic, store.FieldPrvPassTok, prvPassTok)
		if err != nil {
			log.Printf("failed keeping prv pass on short <%s> metadata\n", res[store.FieldPublic])
		}
	}
}
