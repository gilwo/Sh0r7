package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func _spawnErr(c *gin.Context, err error) {
	fmt.Printf("err : %v\n", err)
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

var (
	runningCount = 0
)

func handleCreateShortModDelete(data string, isUrl bool) (map[string]interface{}, error) {
	var err error
	res := map[string]interface{}{}
	shorts := map[string]string{
		"":  shortener.GenerateShortData(data),
		"d": shortener.GenerateShortData(data + "delete"),
		"m": shortener.GenerateShortData(data + "modify"),
	}
	for _, e := range shorts {
		if e == "" {
			return nil, fmt.Errorf("there was a problem creating a short")
		}
	}
	mapping := map[string]string{
		"":  data,
		"d": shorts[""],
		"m": shorts[""],
	}
	if isUrl {
		shorts["url"] = shorts[""]
		mapping["url"] = shorts[""]
	}
	for k, e := range shorts {
		err = store.StoreCtx.SaveDataMapping([]byte(mapping[k]), e+k)
		if err != nil {
			break
		}
	}
	if err == nil {
		for k, e := range shorts {
			if k == "" {
				continue
			}
			err = store.StoreCtx.SetMetaDataMapping(shorts[""], k, e+k)
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
	res["modify"] = shorts["m"]
	res["delete"] = shorts["d"]
	return res, nil
}
func HandleCreateShortData(c *gin.Context) {
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	res, err := handleCreateShortModDelete(string(d), false)
	if err != nil {
		_spawnErr(c, err)
		return
	}
	fmt.Printf("res: %#v\n", res)
	c.JSON(200, res)
}
func HandleCreateShortUrl(c *gin.Context) {
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	mapping := map[string]string{}
	err = json.Unmarshal(d, &mapping)
	if err != nil {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		return
	}
	if url, ok := mapping["url"]; !ok {
		_spawnErr(c, fmt.Errorf("invalid payload"))
		return
	} else {
		res, err := handleCreateShortModDelete(url, true)
		if err != nil {
			_spawnErr(c, err)
			return
		}
		fmt.Printf("res: %#v\n", res)
		c.JSON(200, res)
	}
}
func HandleCreateShortDataImproved1(c *gin.Context) {
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}

	shortValue := shortener.GenerateShortData(string(d))
	if shortValue == "" {
		_spawnErr(c, fmt.Errorf("there was a problem creating a short url, try again shortly"))
		return
	}
	err = store.StoreCtx.SaveDataMapping(d, shortValue)
	if err != nil {
		_spawnErr(c, err)
		return
	}

	res := gin.H{
		"shortData": shortValue,
	}
	token := shortener.GenerateToken(string(d)+shortValue+c.ClientIP(), 22)
	if token == "" {
		fmt.Printf("failed to _generate_ token to data mapping at %s\n", shortValue)
	} else {
		err = store.StoreCtx.SetMetaDataMapping(shortValue, "token", token)
		if err != nil {
			fmt.Printf("failed to _set_ token to data mapping at %s err: %s\n", shortValue, errors.Unwrap(err))
		} else {
			res["token"] = token
		}
	}

	c.JSON(200, res)
}
func HandleUpdateShort(c *gin.Context) {
	short := c.Param("short")
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
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

	mod, err := store.StoreCtx.LoadDataMapping(short + "m")
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
		return false
	}
	url, ok := mapping["url"]
	if !ok {
		_spawnErr(c, fmt.Errorf("invalid payload"))
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
	_, err := store.StoreCtx.LoadDataMapping(short + "m")
	if err != nil {
		fmt.Println(err)
		return false
	}

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
func DeleteShortData(c *gin.Context) {
	short := c.Param("short")
	removeKey := short + "d"
	dataKey, err := store.StoreCtx.LoadDataMapping(removeKey)
	if err != nil {
		_spawnErr(c, err)
		return
	}
	modifyKey, err := store.StoreCtx.GetMetaDataMapping(string(dataKey), "m")
	if err != nil {
		_spawnErr(c, err)
		return
	}

	if err := store.StoreCtx.RemoveDataMapping(string(dataKey)); err != nil {
		_spawnErr(c, err)
		return
	}
	if err := store.StoreCtx.RemoveDataMapping(modifyKey); err != nil {
		_spawnErr(c, err)
		return
	}
	if err := store.StoreCtx.RemoveDataMapping(removeKey); err != nil {
		_spawnErr(c, err)
		return
	}
	store.StoreCtx.RemoveDataMapping(string(dataKey) + "url")
	c.Status(200)
}
func CreateShortData(c *gin.Context) {
	runningCount += 1
	dataMap := make(map[string]interface{})

	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}
	// fmt.Printf("raw data: %+#v\n", d)
	err = json.Unmarshal(d, &dataMap)
	if err != nil {
		_spawnErr(c, err)
		return
	}
	fmt.Printf("raw as json: %#+v\n", dataMap)
	// for k, v := range data.Data {
	// fmt.Printf("post data: %#+v\n")
	// }
	fmt.Printf("from: %v\n", c.Request.RemoteAddr)
	// store.StoreCtx.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId)
	err = store.StoreCtx.SaveDataMapping(d, fmt.Sprintf("%d", runningCount))
	if err != nil {
		_spawnErr(c, err)
		return
	}

	c.JSON(200, gin.H{
		"shortData": runningCount,
	})
}

func HandleGetShortDataInfo(c *gin.Context) {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMappingInfo(short)
	if err != nil {
		_spawnErr(c, err)
		return
	}
	c.JSON(200, data)
}
func HandleGetOriginData(c *gin.Context) {
	shortUrl := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(shortUrl)
	if err != nil {
		_spawnErr(c, err)
		return
	}

	c.String(200, "%s", data)
}

func HandleShort(c *gin.Context) {
	short := c.Param("short")
	fmt.Println("trying url for ", short)
	if tryUrl(c) {
		return
	}
	fmt.Println("trying data for ", short)
	if getData(c) {

		return
	}

	_spawnErr(c, fmt.Errorf("short %s not found", short))
}
func getData(c *gin.Context) bool {
	short := c.Param("short")
	data, err := store.StoreCtx.LoadDataMapping(short)
	if err != nil {
		fmt.Println(err)
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
		return false
	}

	data, err = store.StoreCtx.LoadDataMapping(string(data))
	if err != nil {
		fmt.Println(err)
		return false
	}
	url := string(data)
	if !strings.HasPrefix(url, "http://") || !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	c.Redirect(302, url)
	return true
}
