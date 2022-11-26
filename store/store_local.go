package store

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	entry.(*stringTuple).Set(fmt.Sprintf("changed_time_%d", countNumber), time.Now().Format(time.RFC3339))
	entry.(*stringTuple).Set("data", k)
	// tup, err := NewStringTuple([]*fieldValue{{"data", k}}...)
	// if err != nil {
	// 	return err
	// }
	// storeService.cache[short] = tup
	return nil
}

func (st *StorageLocal) SaveDataMapping(data []byte, short string, ttl time.Duration) error {
	if _, ok := st.cacheSync.Load(short); ok {
		return fmt.Errorf("entry exist for %s", short)
	}
	t := NewTuple()
	err := t.Set2Bytes("data", data, true)
	if err != nil {
		return err
	}
	t.Set("created", time.Now().Format(time.RFC3339))
	if ttl == 0 {
		t.Set("ttl", DefaultExpireDuration.String())
	} else if ttl > 0 {
		t.Set("ttl", ttl.String())
	} // ttl < 0 - dont use ttl at all

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

func (st *StorageLocal) GenFunc(v ...interface{}) interface{} {
	fmt.Printf("!!!!!!!!!! genfunc ... args: <%#v>\n", v)
	if len(v) < 1 {
		return nil
	}
	switch v[0].(string) {
	case STORE_FUNC_GETKEYS:
		fmt.Println("!!!!!!!!!! getkeys ... ")
		return st.getKeys()
	case STORE_FUNC_REMOVEKEYS:
		fmt.Println("!!!!!!!!!! getkeys ... ")
		if len(v) < 2 {
			return nil
		}
		ks := v[1].([]string)
		return st.removeKeys(ks)
	}
	return nil
}

func (st *StorageLocal) getKeys() []string {
	r := []string{}
	st.cacheSync.Range(func(key, value any) bool {
		r = append(r, key.(string))
		return true
	})
	return r
}

func (st *StorageLocal) removeKeys(ks []string) []error {
	fmt.Printf("** removing keys: %#v\n", ks)
	errors := []error{}
	for _, k := range ks {
		if err := st.RemoveDataMapping(k); err != nil {
			errors = append(errors, err)
		}
	}

	errs := []string{}
	for _, e := range errors {
		errs = append(errs, e.Error())
	}
	fmt.Printf("** errors gathered: %#+v\n", strings.Join(errs, "; "))
	return errors
}
