package store

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
// storeService = &StorageService{}
// ctx          = context.Background()
)

// const CacheDuration = 6 * time.Hour

type StorageLocal struct {
	// redisClient *redis.Client
	cacheSync *sync.Map
}

func init() {
	NewStoreLocal = newStoreLocal
}
func newStoreLocal() Store {
	return &StorageLocal{}
}

func (st *StorageLocal) InitializeStore() error {
	st.cacheSync = &sync.Map{}
	return nil
}

// if user id was not provided generate one on the fly : case for not logged in users

/* We want to be able to save the mapping between the originalUrl
and the generated shortUrl url
*/
func (st *StorageLocal) SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	// err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, CacheDuration).Err()
	if _, ok := st.cacheSync.Load(shortUrl); ok {
		panic(fmt.Errorf("overwrite %s with %s, %s", shortUrl, originalUrl, userId))
	}
	vs := []*fieldValue{{"url", originalUrl}, {"user", userId}}
	tup, err := NewStringTuple(vs...)

	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
	st.cacheSync.Store(shortUrl, tup)

	fmt.Printf("content of tuple %#v\n", tup)
	fmt.Printf("Saved shortUrl: %s - originalUrl: %s\n", shortUrl, originalUrl)
}

/*
We should be able to retrieve the initial long URL once the short
is provided. This is when users will be calling the shortlink in the
url, so what we need to do here is to retrieve the long url and
think about redirect.
*/
func (st *StorageLocal) RetrieveInitialUrl(shortUrl string) string {
	// result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	tup, ok := st.cacheSync.Load(shortUrl)
	if !ok {
		panic(fmt.Errorf("no entry for shorturl %s", shortUrl))
	}
	result, err := tup.(*stringTuple).AtCheck("url")
	if err != nil {
		panic(fmt.Sprintf("Failed RetrieveInitialUrl url | Error: %v - shortUrl: %s\n", err, shortUrl))
	}
	return result
}

func (st *StorageLocal) UpdateDataMapping(data []byte, short string) error {
	entry, ok := st.cacheSync.Load(short)
	if !ok {
		return fmt.Errorf("entry not exist for %s", short)
	}
	s := base64.StdEncoding.EncodeToString(data)
	var in bytes.Buffer
	b := []byte(s)
	w, err := zlib.NewWriterLevel(&in, zlib.BestCompression)
	if err != nil {
		return err
	}
	w.Write(b)
	w.Close()

	k := base64.StdEncoding.EncodeToString(in.Bytes())
	entry.(*stringTuple).Set("data", k)

	countNumber := 0
	count, err := entry.(*stringTuple).AtCheck("changed")
	if err != nil {
		// never been changed
	} else {
		countNumber, err = strconv.Atoi(count)
		if err != nil {
			return fmt.Errorf("invalid number of changes")
		}
	}
	// keep old data
	entry.(*stringTuple).Set(fmt.Sprintf("data_%d", countNumber), entry.(*stringTuple).Get("data"))
	countNumber += 1
	entry.(*stringTuple).Set("changed", fmt.Sprintf("%d", countNumber))
	entry.(*stringTuple).Set(fmt.Sprintf("changed_time_%d", countNumber), time.Now().String())
	entry.(*stringTuple).Set("data", k)
	// tup, err := NewStringTuple([]*fieldValue{{"data", k}}...)
	// if err != nil {
	// 	return err
	// }
	// storeService.cache[short] = tup
	return nil
}

func (st *StorageLocal) SaveDataMapping(data []byte, short string) error {
	if _, ok := st.cacheSync.Load(short); ok {
		return fmt.Errorf("entry exist for %s", short)
	}
	t := NewTuple()
	err := t.Set2Bytes("data", data, true)
	if err != nil {
		return err
	}
	t.Set("created", time.Now().String())

	return func() error { st.cacheSync.Store(short, t); return nil }()
}
func (st *StorageLocal) CheckShortDataMapping(short string) error {
	if _, ok := st.cacheSync.Load(short); ok {
		return fmt.Errorf("entry exist for %s", short)
	}
	return nil
}
func (st *StorageLocal) LoadDataMapping(short string) ([]byte, error) {
	tup, ok := st.cacheSync.Load(short)
	if !ok {
		return nil, fmt.Errorf("entry not exist for %s", short)
	}
	return tup.(*stringTuple).Get2Bytes("data")
}
func (st *StorageLocal) LoadDataMappingInfo(short string) (map[string]interface{}, error) {
	tup, ok := st.cacheSync.Load(short)
	if !ok {
		return nil, fmt.Errorf("entry not exist for %s", short)
	}
	ret := map[string]interface{}{}
	for k, v := range tup.(*stringTuple).tuple {
		ret[k] = v
	}
	return ret, nil
}

func (st *StorageLocal) SetMetaDataMapping(short, key, value string) error {
	entry, ok := st.cacheSync.Load(short)
	if !ok {
		return fmt.Errorf("entry not exist for %s", short)
	}
	entry.(*stringTuple).Set(key, value)
	return nil
}

func (st *StorageLocal) GetMetaDataMapping(short, key string) (string, error) {
	entry, ok := st.cacheSync.Load(short)
	if !ok {
		return "", fmt.Errorf("entry not exist for %s", short)
	}
	v, err := entry.(*stringTuple).AtCheck(key)
	if err != nil {
		return "", err
	}
	return v, nil
}

func (st *StorageLocal) RemoveDataMapping(short string) error {
	_, ok := st.cacheSync.Load(short)
	if !ok {
		return fmt.Errorf("entry not exist for %s", short)
	}
	st.cacheSync.Delete(short)
	return nil
}
