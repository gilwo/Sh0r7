package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
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
	LengthMinShortFree      = 6
	LengthMaxShortFree      = 10
	LengthMinPrivate        = 10
	LengthMaxPrivate        = 15
	LengthMinDelete         = 10
	LengthMaxDelete         = 15
	LengthNamedMinShortFree = 10
	LengthNamedMaxShortFree = 40
)

func handleCreateShortModDelete(data, namedPublic string, isPrivate, isRemove, isUrl bool, expiration time.Duration) (map[string]string, error) {
	var err error
	shorts := map[string]string{
		store.SuffixPublic: shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixPublic, -1, 0, LengthMinShortFree, LengthMaxShortFree, store.StoreCtx),
	}
	if namedPublic != "" {
		decodedNamedPublic, err := shortener.Base64SE.Decode(namedPublic)
		if err != nil {
			return nil, errors.Errorf("decode of named public <%s> failed: %s", namedPublic, err)
		}
		namedPublic = string(decodedNamedPublic)
		if len(namedPublic) < LengthNamedMinShortFree {
			return nil, errors.Errorf("named short is too short (%d)", len(namedPublic))
		}
		if len(namedPublic) > LengthNamedMaxShortFree {
			return nil, errors.Errorf("named short is too long (%d)", len(namedPublic))
		}
		// TODO: add here check against reserved names (e.g. favicon.ico, /web/logo.jpg)
		shorts[store.SuffixPublic] = shortener.GenerateShortDataTweakedWithStore2NotRandom(
			namedPublic+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
	}
	if isPrivate {
		shorts[store.SuffixPrivate] = shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixPrivate, -1, 0, LengthMinPrivate, LengthMaxPrivate, store.StoreCtx)
	}
	if isRemove {
		shorts[store.SuffixRemove] = shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixRemove, -1, 0, LengthMinDelete, LengthMaxDelete, store.StoreCtx)
	}
	for k, e := range shorts {
		if e == "" {
			log.Printf("problem with crating short for <%s>\n", k)
			return nil, errors.Errorf("there was a problem creating a short")
		}
	}
	mapping := map[string]string{
		store.FieldPublic: data,
	}
	if isPrivate {
		mapping[store.FieldPrivate] = shorts[store.SuffixPrivate]
	}
	if isRemove {
		mapping[store.FieldRemove] = shorts[store.SuffixRemove]
	}
	if isUrl {
		shorts[store.SuffixURL] = shorts[store.SuffixPublic]
		mapping[store.FieldURL] = shorts[store.SuffixPublic]
	}
	if namedPublic != "" {
		mapping[store.FieldNamedPublic] = namedPublic
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

	res := map[string]string{
		store.FieldPublic: shorts[store.SuffixPublic],
	}
	if namedPublic != "" {
		res[store.FieldPublic] = namedPublic
	}
	if isPrivate {
		res[store.FieldPrivate] = shorts[store.SuffixPrivate]
	}
	if isRemove {
		res[store.FieldRemove] = shorts[store.SuffixRemove]
	}
	return res, nil
}

// TODO: add proper cleanup on failed creation (for any reason... )
func HandleCreateShortData(c *gin.Context) {
	if !checkToken(c) {
		return
	}
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	res, err := handleCreateShortModDelete(string(d),
		c.Request.Header.Get(common.FNamedPublic),
		c.Request.Header.Get(common.FPrivate) != "false",
		c.Request.Header.Get(common.FRemove) != "false",
		false, getExpiration(c))
	if err != nil {
		_spawnErr(c, err)
		return
	}
	err = handleCreateHeaders(c, res)
	if err != nil {
		_spawnErr(c, err)
		return
	}
	log.Printf("res: %#v\n", verboseShorts(res))
	c.JSON(200, res)
}

func verboseShorts(z map[string]string) (r string) {
	for k, v := range z {
		r += fmt.Sprintf("%s: <%s>(%d)\n", k, v, len(v))
	}
	return
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

	// this overwrite - remove and save ... - FIXME - try use updatedatamapping
	// there is a potential of name hijacking in some rare occasions
	store.StoreCtx.RemoveDataMapping(name)
	store.StoreCtx.SaveDataMapping(buf.Bytes(), name, -1)
}

// TODO: add proper cleanup on failed creation (for any reason... )
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
		res, err := handleCreateShortModDelete(url,
			c.Request.Header.Get(common.FNamedPublic),
			c.Request.Header.Get(common.FPrivate) != "false",
			c.Request.Header.Get(common.FRemove) != "false",
			true, getExpiration(c))
		if err != nil {
			_spawnErr(c, err)
			return
		}
		err = handleCreateHeaders(c, res)
		if err != nil {
			_spawnErr(c, err)
			return
		}
		log.Printf("res: %v\n", verboseShorts(res))
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
	var accessAllowed *bool
	if accessAllowed = isAccessToShortAllowed(c, string(dataKey)); accessAllowed != nil && !*accessAllowed {
		return *accessAllowed
	}
	info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic)
	if err != nil {
		log.Printf("failed to get info for token: <%s>\n", dataKey)
		_spawnErrWithCode(c, http.StatusPreconditionFailed, errors.Errorf("invalid token"))
		return false
	}
	if val, ok := info[store.FieldRemove]; ok { // we know its here alread as we arrived from it ... but for the generic flow logic - need to optimize ...
		removeKey := val.(string) + store.SuffixRemove
		if err := store.StoreCtx.RemoveDataMapping(removeKey); err != nil {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s - remove, err: %s\n", msg, err)
			_spawnErrCond(c, msg, withResponse)
			return false
		}
	}
	if val, ok := info[store.FieldPublic]; ok {
		publicKey := val.(string) + store.SuffixPublic
		if err := store.StoreCtx.RemoveDataMapping(publicKey); err != nil {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s - public, err: %s\n", msg, err)
			_spawnErrCond(c, msg, withResponse)
			return false
		}
	}
	if val, ok := info[store.FieldPrivate]; ok {
		privateKey := val.(string) + store.SuffixPrivate
		if err := store.StoreCtx.RemoveDataMapping(privateKey); err != nil {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s - private, err: %s\n", msg, err)
			_spawnErrCond(c, msg, withResponse)
			return false
		}
	}
	if val, ok := info[store.FieldURL]; ok {
		urlKey := val.(string) + store.SuffixURL
		if err := store.StoreCtx.RemoveDataMapping(urlKey); err != nil {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s - url, err: %s\n", msg, err)
			_spawnErrCond(c, msg, withResponse)
			return false
		}
	}
	mt := PrepMetricShortAccess(c, short, true /*success*/, false /*private*/, true /*delete*/, accessAllowed /*locked*/)
	metrics.MetricGlobalCounter.IncShortAccessVisitDeleteCount()
	metrics.MetricProcessor.Add(mt)
	return true
}

func HandleGetShortDataInfo(c *gin.Context) {
	if getDataPrivate(c) {
		return
	}
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

func isAccessToShortAllowed(c *gin.Context, short string) (r *bool) {
	True := true
	False := false
	storedPubPassTok, e1 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, store.FieldPubPassTok)
	if e1 == nil && storedPubPassTok != "" {
		storedPubPassSalt, e2 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, store.FieldPubPassSalt)
		if e2 != nil {
			msg := errors.Errorf("short <%s> is locked but missing salt", short)
			log.Printf("%s, e2: %s\n", msg, e2.Error())
			_spawnErr(c, msg)
			return &False
		}
		if !c.Request.URL.Query().Has(common.FPass) {
			recvPassToken := c.Request.Header.Get(common.FPubPassToken)
			log.Printf("-- (recv) pass token: <%s>\n", recvPassToken)
			log.Printf("-- pass token: <%s>\n", storedPubPassTok)
			log.Printf("-- pass salt: <%s>\n", storedPubPassSalt)
			if recvPassToken != storedPubPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass tok),  recv <%s>, salt <%s>, stored <%s>\n",
					msg, recvPassToken, storedPubPassSalt, storedPubPassTok)
				_spawnErrWithCode(c, http.StatusForbidden, msg)
				return &False
			}
		} else {
			passTxt := c.Request.URL.Query().Get(common.FPass)
			calcPassTok := shortener.GenerateTokenTweaked(passTxt+storedPubPassSalt, 0, 30, 10)
			log.Printf("-- (recv) pass : <%s>\n", passTxt)
			log.Printf("-- pass salt: <%s>\n", storedPubPassSalt)
			log.Printf("-- pass token: <%s>\n", storedPubPassTok)
			log.Printf("-- (calc) pass token: <%s>\n", calcPassTok)
			if calcPassTok != storedPubPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass text), txt <%s>, salt <%s>, calc <%s>, stored <%s>\n",
					msg, passTxt, storedPubPassSalt, calcPassTok, storedPubPassTok)
				_spawnErrWithCode(c, http.StatusForbidden, msg)
				return &False
			}
		}
		return &True
	}
	return
}
func isAccessToRemoveAllowed(c *gin.Context, short string) (r *bool) {
	True := true
	False := false
	storedRemPassTok, e1 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, store.FieldRemPassTok)
	if e1 == nil && storedRemPassTok != "" {
		storedRemPassSalt, e2 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, store.FieldRemPassSalt)
		if e2 != nil {
			msg := errors.Errorf("short <%s> is locked but missing salt", short)
			log.Printf("%s, e2: %s\n", msg, e2.Error())
			_spawnErr(c, msg)
			return &False
		}
		if !c.Request.URL.Query().Has(common.FPass) {
			recvPassToken := c.Request.Header.Get(common.FRemPassToken)
			log.Printf("-- (recv) pass token: <%s>\n", recvPassToken)
			log.Printf("-- pass token: <%s>\n", storedRemPassTok)
			log.Printf("-- pass salt: <%s>\n", storedRemPassSalt)
			if recvPassToken != storedRemPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass tok),  recv <%s>, salt <%s>, stored <%s>\n",
					msg, recvPassToken, storedRemPassSalt, storedRemPassTok)
				_spawnErrWithCode(c, http.StatusForbidden, msg)
				return &False
			}
		} else {
			passTxt := c.Request.URL.Query().Get(common.FPass)
			calcPassTok := shortener.GenerateTokenTweaked(passTxt+storedRemPassSalt, 0, 30, 10)
			log.Printf("-- (recv) pass : <%s>\n", passTxt)
			log.Printf("-- pass salt: <%s>\n", storedRemPassSalt)
			log.Printf("-- pass token: <%s>\n", storedRemPassTok)
			log.Printf("-- (calc) pass token: <%s>\n", calcPassTok)
			if calcPassTok != storedRemPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass text), txt <%s>, salt <%s>, calc <%s>, stored <%s>\n",
					msg, passTxt, storedRemPassSalt, calcPassTok, storedRemPassTok)
				_spawnErrWithCode(c, http.StatusForbidden, msg)
				return &False
			}
		}
		return &True
	}
	return
}
func getData(c *gin.Context) bool {
	short := c.Param("short")
	if !store.StoreCtx.CheckExistShortDataMapping(short + store.SuffixPublic) {
		shortNamed := shortener.GenerateShortDataTweakedWithStore2NotRandom(short+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
		if !store.StoreCtx.CheckExistShortDataMapping(shortNamed + store.SuffixPublic) {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s, not found when getting data - also for named public option (%s)\n", msg, shortNamed)
			return false
		}
		short = string(shortNamed)
	}
	var accessAllowed *bool
	if accessAllowed = isAccessToShortAllowed(c, string(short)); accessAllowed != nil && !*accessAllowed {
		return *accessAllowed
	}
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
	mt := PrepMetricShortAccess(c, short, true /*success*/, false /*private*/, false /*delete*/, accessAllowed /*locked*/)
	metrics.MetricGlobalCounter.IncShortAccessVisitCount()
	metrics.MetricProcessor.Add(mt)
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
	var accessAllowed *bool
	if accessAllowed = isAccessToPrivateAllowed(c, string(dataKey)); accessAllowed != nil && !*accessAllowed {
		return *accessAllowed
	}
	data, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic)
	if err != nil {
		msg := errors.Errorf("there was a problem with short: %s", short)
		log.Printf("%s, err: %s\n", msg, err)
		return false
	}
	c.JSON(200, data)
	mt := PrepMetricShortAccess(c, short, false /*success*/, true /*private*/, false /*delete*/, accessAllowed /*locked*/)
	metrics.MetricGlobalCounter.IncShortAccessVisitCount()
	metrics.MetricProcessor.Add(mt)
	return true
}

func tryUrl(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short + store.SuffixURL)
	if err != nil {
		shortNamed := shortener.GenerateShortDataTweakedWithStore2NotRandom(short+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
		data, err = store.StoreCtx.LoadDataMapping(shortNamed + store.SuffixURL)
		if err != nil {
			log.Println(err)
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("load url: %s, err: %s\n", msg, err)
			return false
		}
	}
	var accessAllowed *bool
	if accessAllowed = isAccessToShortAllowed(c, string(data)); accessAllowed != nil && !*accessAllowed {
		return *accessAllowed
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
	if c.Request.Header.Get("xRedirect") == "no" {
		tup := store.NewTuple()
		tup.Set(store.FieldURL, url)
		c.String(200, "%s", tup.ToString())
		// c.JSON(200, map[string]interface{}{store.FieldURL: url})
	} else {
		c.Redirect(302, url)
	}
	mt := PrepMetricShortAccess(c, short, true /*success*/, false /*private*/, false /*delete*/, accessAllowed /*locked*/)
	metrics.MetricGlobalCounter.IncShortAccessVisitCount()
	metrics.MetricProcessor.Add(mt)
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

	v, ok := info[store.FieldTime]
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
	v, ok = info[store.FieldTTL]
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

func handleCreateHeaders(c *gin.Context, res map[string]string) error {
	var err error
	short := res[store.FieldPublic]
	if !store.StoreCtx.CheckExistShortDataMapping(short + store.SuffixPublic) {
		shortNamed := shortener.GenerateShortDataTweakedWithStore2NotRandom(short+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
		if !store.StoreCtx.CheckExistShortDataMapping(shortNamed + store.SuffixPublic) {
			msg := errors.Errorf("there was a problem with short: %s", short)
			log.Printf("%s, not found when getting data - also for named public option (%s)\n", msg, shortNamed)
			return errors.Errorf("failed to set password for public link")
		}
		short = string(shortNamed)
	}
	if desc := c.Request.Header.Get(common.FShortDesc); desc != "" {
		store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldDesc, desc)
	}
	if prvPassTok := c.Request.Header.Get(common.FPrvPassToken); prvPassTok != "" {
		token := c.Request.Header.Get(common.FTokenID)
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldPrvPassSalt, token)
		if err != nil {
			log.Printf("failed keeping prv pass token on short <%s> metadata\n", short)
			return errors.Errorf("failed to set password for private link")
		}
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldPrvPassTok, prvPassTok)
		if err != nil {
			log.Printf("failed keeping prv pass on short <%s> metadata\n", short)
			return errors.Errorf("failed to set password for private link")
		}
	}
	if pubPassTok := c.Request.Header.Get(common.FPubPassToken); pubPassTok != "" {
		token := c.Request.Header.Get(common.FTokenID)
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldPubPassSalt, token)
		if err != nil {
			log.Printf("failed keeping pub pass token on short <%s> metadata : %s\n", short, err)
			return errors.Errorf("failed to set password for public link")
		}
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldPubPassTok, pubPassTok)
		if err != nil {
			log.Printf("failed keeping pub pass on short <%s> metadata : %s\n", short, err)
			return errors.Errorf("failed to set password for public link")
		}
	}
	if remPassTok := c.Request.Header.Get(common.FRemPassToken); remPassTok != "" {
		token := c.Request.Header.Get(common.FTokenID)
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldRemPassSalt, token)
		if err != nil {
			log.Printf("failed keeping rem pass token on short <%s> metadata\n", short)
			return errors.Errorf("failed to set password for remove link")
		}
		err = store.StoreCtx.SetMetaDataMapping(short+store.SuffixPublic, store.FieldRemPassTok, remPassTok)
		if err != nil {
			log.Printf("failed keeping rem pass on short <%s> metadata\n", short)
			return errors.Errorf("failed to set password for remove link")
		}
	}
	return nil
}

func HandleDumpKeys(c *gin.Context) {
	res := store.StoreCtx.GenFunc(store.STORE_FUNC_DUMPALL).(string)
	log.Println("!! dumpall: \n" + res)
	c.String(200, "%s", res)
}

func PrepMetricShortAccess(c *gin.Context, name string, success, private, delete bool, isLocked *bool) *metrics.MetricShortAccess {
	mt := metrics.NewMetricShortAccess()
	mt.ShortAccessVisitName = name
	mt.ShortAccessVisitIP = c.ClientIP()
	mt.ShortAccessVisitTime = time.Now().String()
	// mt.ShortAccessVisitSuccess = fmt.Sprintf("%t", success)
	mt.ShortAccessVisitSuccess = func() string {
		if success {
			return "true"
		}
		return "false"
	}()
	// mt.ShortAccessVisitPrivate = fmt.Sprintf("%t", private)
	mt.ShortAccessVisitPrivate = func() string {
		if private {
			return "true"
		}
		return "false"
	}()
	// mt.ShortAccessVisitDelete = fmt.Sprintf("%t", delete)
	mt.ShortAccessVisitDelete = func() string {
		if delete {
			return "true"
		}
		return "false"
	}()
	mt.ShortAccessVisitIsLocked = func() string {
		if isLocked == nil {
			return "false"
		}
		return "true"
	}()
	reqDump, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		log.Printf("failed getting request dump for %s, err: %s\n", name, err)
		reqDump = []byte("something failed getting request dump: " + err.Error())
	}
	mt.ShortAccessVisitInfo = fmt.Sprintf("\n--\n%s\n--\n%s\n--\n%s",
		c.Request.RemoteAddr,
		c.Request.RequestURI,
		string(reqDump),
	)
	return mt
}

