package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gilwo/Sh0r7/metrics"
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	"github.com/gilwo/Sh0r7/webapp/common"
	"github.com/gin-gonic/gin"
	"github.com/karrick/tparse"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	LengthMinRemove         = 10
	LengthMaxRemove         = 15
	LengthNamedMinShortFree = 10
	LengthNamedMaxShortFree = 40
)

func handleCreateShortModRemove(data, namedPublic string, isPrivate, isRemove, isUrl bool, expiration time.Duration) (shorts map[string]string, err error) {
	shorts = map[string]string{
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
		if checkReserveNames(namedPublic).IsFalse() {
			return nil, errors.Errorf("named short (%s) cannot be used", namedPublic)
		}
		shorts[store.SuffixPublic] = shortener.GenerateShortDataTweakedWithStore2NotRandom(
			namedPublic+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
	}
	if isPrivate {
		shorts[store.SuffixPrivate] = shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixPrivate, -1, 0, LengthMinPrivate, LengthMaxPrivate, store.StoreCtx)
	}
	if isRemove {
		shorts[store.SuffixRemove] = shortener.GenerateShortDataTweakedWithStore2(
			data+store.SuffixRemove, -1, 0, LengthMinRemove, LengthMaxRemove, store.StoreCtx)
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
	fail := false
	tokenFail := false
	defer func() {
		if fail {
			if tokenFail {
				metrics.GlobalMeter.IncMeterCounter(metrics.CreationFailedToken)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.CreationFailed)
			}
		} else {
			metrics.GlobalMeter.IncMeterCounter(metrics.Created)
		}
	}()
	if !checkToken(c) {
		fail = true
		tokenFail = true
		return
	}
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		fail = true
		return
	}
	res, err := handleCreateShortModRemove(string(d),
		c.Request.Header.Get(common.FNamedPublic),
		c.Request.Header.Get(common.FPrivate) != "false",
		c.Request.Header.Get(common.FRemove) != "false",
		false, getExpiration(c))
	if err != nil {
		fail = true
		_spawnErr(c, err)
		return
	}
	err = handleCreateHeaders(c, res)
	if err != nil {
		fail = true
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
	fail := false
	tokenFail := false
	defer func() {
		if fail {
			if tokenFail {
				metrics.GlobalMeter.IncMeterCounter(metrics.CreationFailedToken)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.CreationFailed)
			}
		} else {
			metrics.GlobalMeter.IncMeterCounter(metrics.Created)
		}
	}()
	if !checkToken(c) {
		fail = true
		tokenFail = true
		return
	}
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		fail = true
		return
	}
	mapping := map[string]string{}
	err = json.Unmarshal(d, &mapping)
	if err != nil {
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("err in json unmarshal, %s\n", err)
		fail = true
		return
	}
	if url, ok := mapping["url"]; !ok {
		_spawnErr(c, errors.Errorf("invalid payload"))
		log.Printf("json field url is missing, %#v\n", mapping)
		fail = true
		return
	} else {
		res, err := handleCreateShortModRemove(url,
			c.Request.Header.Get(common.FNamedPublic),
			c.Request.Header.Get(common.FPrivate) != "false",
			c.Request.Header.Get(common.FRemove) != "false",
			true, getExpiration(c))
		if err != nil {
			_spawnErr(c, err)
			fail = true
			return
		}
		err = handleCreateHeaders(c, res)
		if err != nil {
			_spawnErr(c, err)
			fail = true
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
func RemoveShortData(c *gin.Context) {
	if handleRemove(c).IsNil() {
		_spawnErrWithCode(c, http.StatusNotFound, errors.New(c.Request.URL.Path+" not found"))
		metrics.GlobalMeter.IncMeterCounter(metrics.InvalidShort)
	}
}

// handleRemove:
//
//	True - removed succeeded - update response
//	False - removed failed due to an error or failure to unlock - update response
//	Nil - removed unhandled (not found) - skipped respones
func handleRemove(c *gin.Context) (r resTri) {
	defer func() {
		if r.IsTrue() {
			metrics.GlobalMeter.IncMeterCounter(metrics.ShortRemoved)
		} else if r.IsFalse() {
			if c.Writer.Status() != http.StatusUnauthorized {
				metrics.GlobalMeter.IncMeterCounter(metrics.RemoveFailed)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.RemoveNotAuth)
			}
		}
	}()
	r = ResTri()
	short := c.Param("short")
	removeKey := short + store.SuffixRemove

	if !store.StoreCtx.CheckExistShortDataMapping(removeKey) {
		return r.Nil()
	}

	errMsg := errors.Errorf("there was a problem with short: %s", short)
	dataKey, err := store.StoreCtx.LoadDataMapping(removeKey)
	if err != nil {
		log.Printf("failed getting public from remove key: <%s>, err: %s\n", removeKey, err)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}
	accessRes := shortAccessAllowedCheck(c, string(dataKey), common.ShortRemove)
	if accessRes.IsFalse() {
		return r.False()
	}
	info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic)
	if err != nil {
		log.Printf("failed to get info for public: <%s>\n", dataKey)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}

	var removeErr error
	// remove all even if some fail , but log it
	if val, ok := info[store.FieldRemove]; ok { // we know its here alread as we arrived from it ... but for the generic flow logic - need to optimize ...
		removeKey := val.(string) + store.SuffixRemove
		if removeErr = store.StoreCtx.RemoveDataMapping(removeKey); removeErr != nil {
			log.Printf("removal of remove key <%s>, err: %s\n", removeKey, removeErr)
		}
	}
	if val, ok := info[store.FieldPublic]; ok {
		publicKey := val.(string) + store.SuffixPublic
		if removeErr = store.StoreCtx.RemoveDataMapping(publicKey); removeErr != nil {
			log.Printf("removal of public key <%s>, err: %s\n", publicKey, removeErr)
		}
	}
	if val, ok := info[store.FieldPrivate]; ok {
		privateKey := val.(string) + store.SuffixPrivate
		if removeErr = store.StoreCtx.RemoveDataMapping(privateKey); removeErr != nil {
			log.Printf("removal of private key <%s>, err: %s\n", privateKey, removeErr)
		}
	}
	if val, ok := info[store.FieldURL]; ok {
		urlKey := val.(string) + store.SuffixURL
		if removeErr = store.StoreCtx.RemoveDataMapping(urlKey); removeErr != nil {
			log.Printf("removal of url key <%s>, err: %s\n", urlKey, removeErr)
		}
	}
	if removeErr != nil {
		_spawnErrWithCode(c, http.StatusInternalServerError, errors.New("failed to remove some keys"))
		return r.False()
	}
	c.Status(200)
	return r.True()
}

func HandleGetShortDataInfo(c *gin.Context) {
	if handlePrivateData(c).IsNil() {
		path := c.Request.URL.Path
		_spawnErrWithCode(c, http.StatusNotFound, errors.New(path+" not found"))
		metrics.GlobalMeter.IncMeterCounter(metrics.InvalidShort)
	}
}

func HandleGetOriginData(c *gin.Context) {
	if handleData(c).IsNil() {
		path := c.Request.URL.Path
		_spawnErrWithCode(c, http.StatusNotFound, errors.New(path+" not found"))
		metrics.GlobalMeter.IncMeterCounter(metrics.InvalidShort)
	}
}

func HandleShort(c *gin.Context) {
	short := c.Param("short")
	log.Printf("trying url for <%s>\n", short)

	if !handleUrl(c).IsNil() {
		{
			span := trace.SpanFromContext(c.Request.Context())
			if span.IsRecording() {
				span.SetAttributes(attribute.String("referrer", c.Request.Referer()))
				span.SetAttributes(attribute.String("which", common.ShortPublic.String()))
				span.SetAttributes(attribute.String("key", short))
				span.SetAttributes(attribute.Bool("url", true))
			}
		}
		return
	}
	log.Printf("trying data for <%s>\n", short)
	if !handleData(c).IsNil() {
		{
			span := trace.SpanFromContext(c.Request.Context())
			if span.IsRecording() {
				span.SetAttributes(attribute.String("referrer", c.Request.Referer()))
				span.SetAttributes(attribute.String("which", common.ShortPublic.String()))
				span.SetAttributes(attribute.String("key", short))
			}
		}
		return
	}
	log.Printf("trying private data for <%s>\n", short)
	if !handlePrivateData(c).IsNil() {
		{
			span := trace.SpanFromContext(c.Request.Context())
			if span.IsRecording() {
				span.SetAttributes(attribute.String("referrer", c.Request.Referer()))
				span.SetAttributes(attribute.String("which", common.ShortPrivate.String()))
				span.SetAttributes(attribute.String("key", short))
			}
		}
		return
	}
	log.Printf("trying remove for <%s>\n", short)
	if !handleRemove(c).IsNil() {
		{
			span := trace.SpanFromContext(c.Request.Context())
			if span.IsRecording() {
				span.SetAttributes(attribute.String("referrer", c.Request.Referer()))
				span.SetAttributes(attribute.String("which", common.ShortRemove.String()))
				span.SetAttributes(attribute.String("key", short))
			}
		}
		return
	}

	_spawnErrWithCode(c, http.StatusNotFound, errors.Errorf("short %s not found", short))
	metrics.GlobalMeter.IncMeterCounter(metrics.InvalidShort)
}

// handleData: retrieve data associated with short
//
//	True - found - update response
//	False - data access failed due to an error or failure to unlock - update response
//	Nil - unhandled (not found) - skipped respones
func handleData(c *gin.Context) (r resTri) {
	defer func() {
		if r.IsTrue() {
			metrics.GlobalMeter.IncMeterCounter(metrics.VisitPublic)
		} else if r.IsFalse() {
			if c.Writer.Status() != http.StatusUnauthorized {
				metrics.GlobalMeter.IncMeterCounter(metrics.PublicFailed)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.PublicNotAuth)
			}
		}
	}()
	r = ResTri()
	short := c.Param("short")
	dataKey := short
	if !store.StoreCtx.CheckExistShortDataMapping(short + store.SuffixPublic) {
		shortNamed := shortener.GenerateShortDataTweakedWithStore2NotRandom(short+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
		if !store.StoreCtx.CheckExistShortDataMapping(shortNamed + store.SuffixPublic) {
			{ // a hacky flow - TODO - investigate or rethink
				// if short fail here, then try to get the data for the full path
				//  (some elements are stored this way using the storage provider)
				data, err := store.StoreCtx.LoadDataMapping(c.Request.URL.Path)
				if err == nil {
					c.Data(200, "", data)
					return r.True()
				}
			}
			return r.Nil()
		}
		dataKey = shortNamed
	}

	accessRes := shortAccessAllowedCheck(c, string(dataKey), common.ShortPublic)
	if accessRes.IsFalse() {
		return r.False()
	}

	errMsg := errors.Errorf("there was a problem with short: %s", short)
	data, err := store.StoreCtx.LoadDataMapping(string(dataKey) + store.SuffixPublic)
	if err != nil {
		log.Printf("failed to get info for public: <%s> - %s\n", dataKey, err)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}
	c.String(200, "%s", data)

	return r.True()
}

// handlePrivateData: retrieve private data associated with short
//
//	True - found - update response
//	False - data access failed due to an error or failure to unlock - update response
//	Nil - unhandled (not found) - skipped respones
func handlePrivateData(c *gin.Context) (r resTri) {
	defer func() {
		if r.IsTrue() {
			metrics.GlobalMeter.IncMeterCounter(metrics.VisitPrivate)
		} else if r.IsFalse() {
			if c.Writer.Status() != http.StatusUnauthorized {
				metrics.GlobalMeter.IncMeterCounter(metrics.PrivateFailed)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.PrivateNotAuth)
			}
		}
	}()
	r = ResTri()
	short := c.Param("short")
	privateKey := short + store.SuffixPrivate
	if !store.StoreCtx.CheckExistShortDataMapping(privateKey) {
		return r.Nil()
	}
	errMsg := errors.Errorf("there was a problem with short: %s", short)
	dataKey, err := store.StoreCtx.LoadDataMapping(privateKey)
	if err != nil {
		log.Printf("failed getting public from private key: <%s>, err: %s\n", privateKey, err)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}

	accessRes := shortAccessAllowedCheck(c, string(dataKey), common.ShortPrivate)
	if accessRes.IsFalse() {
		return r.False()
	}

	info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic)
	if err != nil {
		log.Printf("failed to get info for public: <%s> - %s\n", dataKey, err)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}
	c.JSON(200, info)

	return r.True()
}

// handleUrl:
//
//	True - found - update response
//	False - url access failed due to an error or failure to unlock - update response
//	Nil - unhandled (not found) - skipped respones
func handleUrl(c *gin.Context) (r resTri) {
	defer func() {
		if r.IsTrue() {
			metrics.GlobalMeter.IncMeterCounter(metrics.VisitPublic)
		} else if r.IsFalse() {
			if c.Writer.Status() != http.StatusUnauthorized {
				metrics.GlobalMeter.IncMeterCounter(metrics.PublicFailed)
			} else {
				metrics.GlobalMeter.IncMeterCounter(metrics.PublicNotAuth)
			}
		}
	}()
	r = ResTri()
	short := c.Param("short")

	shortNamed := shortener.GenerateShortDataTweakedWithStore2NotRandom(short+store.SuffixPublic, 0, common.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
	if !store.StoreCtx.CheckExistShortDataMapping(short+store.SuffixURL) &&
		!store.StoreCtx.CheckExistShortDataMapping(shortNamed+store.SuffixURL) {
		return r.Nil()
	}

	errMsg := errors.Errorf("there was a problem with short: %s", short)
	dataKey, err := store.StoreCtx.LoadDataMapping(short + store.SuffixURL)
	if err != nil {
		dataKey, err = store.StoreCtx.LoadDataMapping(shortNamed + store.SuffixURL)
		if err != nil {
			log.Printf("problem with loading url for short:(named): <%s>:<%s>\n", short, shortNamed)
			_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
			return r.False()
		}
		log.Printf("url mapping for short: <%s>, found as named short<%s\n", short, shortNamed)
	}
	accessRes := shortAccessAllowedCheck(c, string(dataKey), common.ShortPublic)
	if accessRes.IsFalse() {
		return r.False()
	}
	data, err := store.StoreCtx.LoadDataMapping(string(dataKey) + store.SuffixPublic)
	if err != nil {
		log.Printf("failed to get info for public: <%s>\n", dataKey)
		_spawnErrWithCode(c, http.StatusInternalServerError, errMsg)
		return r.False()
	}
	url := string(data)
	if c.Request.Header.Get("xRedirect") == "no" {
		tup := store.NewTuple()
		tup.Set(store.FieldURL, url)
		c.String(200, "%s", tup.ToString())
	} else {
		c.Redirect(302, url)
	}

	return r.True()
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

// TODO: validate legitimate expiration and override or fail ? (meed to choose)
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

// shortAccessAllowedCheck :
//
//	True - access locked and unlock succeeded
//	False - access locked and unlock failed or error occurred - response updated
//	Nil - access is not locked
func shortAccessAllowedCheck(c *gin.Context, short string, which common.ShortType) (r resTri) {
	r = ResTri()
	var (
		tokenHeader string
		tokenField  string
		saltField   string
	)
	switch which {
	case common.ShortPublic:
		tokenHeader = common.FPubPassToken
		tokenField = store.FieldPubPassTok
		saltField = store.FieldPubPassSalt
	case common.ShortPrivate:
		tokenHeader = common.FPrvPassToken
		tokenField = store.FieldPrvPassTok
		saltField = store.FieldPrvPassSalt
	case common.ShortRemove:
		tokenHeader = common.FRemPassToken
		tokenField = store.FieldRemPassTok
		saltField = store.FieldRemPassSalt
	}
	storedPassTok, e1 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, tokenField)
	if e1 == nil && storedPassTok != "" {
		storedPassSalt, e2 := store.StoreCtx.GetMetaDataMapping(string(short)+store.SuffixPublic, saltField)
		if e2 != nil {
			msg := errors.Errorf("short <%s> is locked but missing salt", short)
			log.Printf("%s, e2: %s\n", msg, e2.Error())
			_spawnErrWithCode(c, http.StatusForbidden, msg)
			return r.False()
		}
		if !c.Request.URL.Query().Has(common.FPass) {
			recvPassToken := c.Request.Header.Get(tokenHeader)
			log.Printf("-- (recv) pass token: <%s>\n", recvPassToken)
			log.Printf("-- pass token: <%s>\n", storedPassTok)
			log.Printf("-- pass salt: <%s>\n", storedPassSalt)
			if recvPassToken != storedPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass tok),  recv <%s>, salt <%s>, stored <%s>\n",
					msg, recvPassToken, storedPassSalt, storedPassTok)
				_spawnErrWithCode(c, http.StatusUnauthorized, msg)
				return r.False()
			}
		} else {
			passTxt := c.Request.URL.Query().Get(common.FPass)
			calcPassTok := shortener.GenerateTokenTweaked(passTxt+storedPassSalt, 0, 30, 10)
			log.Printf("-- (recv) pass : <%s>\n", passTxt)
			log.Printf("-- pass salt: <%s>\n", storedPassSalt)
			log.Printf("-- pass token: <%s>\n", storedPassTok)
			log.Printf("-- (calc) pass token: <%s>\n", calcPassTok)
			if calcPassTok != storedPassTok {
				msg := errors.Errorf("access denied to short <%s>", short)
				log.Printf("%s (pass text), txt <%s>, salt <%s>, calc <%s>, stored <%s>\n",
					msg, passTxt, storedPassSalt, calcPassTok, storedPassTok)
				_spawnErrWithCode(c, http.StatusUnauthorized, msg)
				return r.False()
			}
		}
		return r.True()
	}
	return r.Nil()
}

func collectInfo(c *gin.Context, name string) string {
	reqDump, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		log.Printf("failed getting request dump for %s, err: %s\n", name, err)
		reqDump = []byte("something failed getting request dump: " + err.Error())
	}
	info := map[string]string{
		"remote_addr":  c.Request.RemoteAddr,
		"request_uri":  c.Request.RequestURI,
		"request_dump": string(reqDump),
	}
	res, err := json.Marshal(info)
	if err != nil {
		log.Printf("failed converting info to json %#+v, err: %s\n", info, err)
	}
	return string(res)
}

func checkReserveNames(name string) resTri {
	r := ResTri()

	switch name {
	case "web", "favicon.ico", "admin", "create-short-date",
		"create-short-url", "app.css", "wasm_exec.js", "app.js",
		"dump",
		common.ShortPath,
		common.PrivatePath,
		common.PublicPath,
		common.RemovePath:
		return r.False()
	}

	return r.True()
}
