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

// const CacheDuration = 6 * time.Hour

type StorageLocal struct {
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
