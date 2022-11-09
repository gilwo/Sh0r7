package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

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
func UpdateShortData(c *gin.Context) {
	short := c.Param("short")
	d, err := c.GetRawData()
	if err != nil {
		_spawnErr(c, err)
		return
	}

	err = store.StoreCtx.UpdateDataMapping(d, short)
	if err != nil {
		_spawnErr(c, err)
		return
	}

	c.Status(200)
}
func DeleteShortData(c *gin.Context) {
	short := c.Param("short")
	token, err := store.StoreCtx.GetMetaDataMapping(short, "token")
	if err != nil {
		_spawnErr(c, err)
		return
	}
	if c.GetHeader("token") != token {
		_spawnErr(c, fmt.Errorf("invalid token for short %s", short))
		return
	}
	if err := store.StoreCtx.RemoveDataMapping(short); err != nil {
		_spawnErr(c, err)
		return
	}
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

	c.JSON(200, gin.H{
		"data": string(data),
	})
}

