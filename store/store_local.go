package store

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
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
		return errors.Errorf("entry not exist for %s", short)
	}

	countNumber := 0
	count, err := entry.(*stringTuple).AtCheck(FieldModCount)
	if err != nil {
		// never been changed
	} else {
		countNumber, err = strconv.Atoi(count)
		if err != nil {
			return errors.Errorf("invalid number of changes")
		}
	}
	t := entry.(*stringTuple)
	k, err := t.Get2(FieldDATA)
	if err != nil {
		return errors.Errorf("some problem in original data extraction")
	}
	// keep old data
	t.Set2(fmt.Sprintf("%s_%d", FieldDATA, countNumber), k, true)
	countNumber += 1
	t.Set(FieldModCount, fmt.Sprintf("%d", countNumber))
	t.Set(fmt.Sprintf("%s_%d", FieldModTime, countNumber), time.Now().Format(time.RFC3339))
	t.Set2Bytes(FieldDATA, data, true)
	return nil
}

func (st *StorageLocal) SaveDataMapping(data []byte, short string, ttl time.Duration) error {
	if _, ok := st.cacheSync.Load(short); ok {
		return errors.Errorf("entry exist for %s", short)
	}
	t := NewTuple()
	err := t.Set2Bytes(FieldDATA, data, true)
	if err != nil {
		return err
	}
	t.Set(FieldTime, time.Now().Format(time.RFC3339))
	if ttl == 0 {
		t.Set(FieldTTL, DefaultExpireDuration.String())
	} else if ttl > 0 {
		t.Set(FieldTTL, ttl.String())
	} // ttl < 0 - dont use ttl at all

	return func() error { st.cacheSync.Store(short, t); return nil }()
}
func (st *StorageLocal) CheckExistShortDataMapping(short string) bool {
	if _, ok := st.cacheSync.Load(short); ok {
		return true
	}
	return false
}
func (st *StorageLocal) LoadDataMapping(short string) ([]byte, error) {
	tup, ok := st.cacheSync.Load(short)
	if !ok {
		return nil, errors.Errorf("entry not exist for %s", short)
	}
	t := tup.(*stringTuple)
	if t.Get(FieldBlocked) == IsBLOCKED {
		return nil, errors.Errorf("not allowed %s", short)
	}
	return t.Get2Bytes(FieldDATA)
}
func (st *StorageLocal) LoadDataMappingInfo(short string) (map[string]interface{}, error) {
	tup, ok := st.cacheSync.Load(short)
	if !ok {
		return nil, errors.Errorf("entry not exist for %s", short)
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
		return errors.Errorf("entry not exist for %s", short)
	}
	entry.(*stringTuple).Set(key, value)
	return nil
}

func (st *StorageLocal) GetMetaDataMapping(short, key string) (string, error) {
	entry, ok := st.cacheSync.Load(short)
	if !ok {
		return "", errors.Errorf("entry not exist for %s", short)
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
		return errors.Errorf("entry not exist for %s", short)
	}
	st.cacheSync.Delete(short)
	return nil
}

func (st *StorageLocal) GenFunc(v ...interface{}) interface{} {
	log.Printf("!!!!!!!!!! genfunc ... args: <%#v>\n", v)
	if len(v) < 1 {
		return nil
	}
	switch v[0].(string) {
	case STORE_FUNC_DUMP:
		if len(v) < 2 {
			return nil
		}
		k := v[1].(string)
		return st.dumpKey(k)
	case STORE_FUNC_DUMPALL:
		return st.dumpAll()
	case STORE_FUNC_DUMPKEYS:
		return st.dumpKeys()
	case STORE_FUNC_GETKEYS:
		return st.getKeys()
	case STORE_FUNC_REMOVEKEYS:
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
func (st *StorageLocal) dumpKeys() string {
	r := st.getKeys()
	sort.Strings(r)
	return strings.Join(r, "\n")
}
func (st *StorageLocal) dumpAll() string {
	r := st.getKeys()
	sort.Strings(r)
	res := ""
	for _, k := range r {
		res += k + "\n" + st.dumpKey(k) + "\n\n"
	}
	return res
}
func (st *StorageLocal) dumpKey(k string) string {
	if v, ok := st.cacheSync.Load(k); ok {
		tup := v.(*stringTuple)
		return tup.Dump()
	}
	return "empty"
}

func (st *StorageLocal) removeKeys(ks []string) []error {
	errors := []error{}
	if len(ks) == 0 {
		return errors
	}
	log.Printf("** removing keys: %#v\n", ks)
	for _, k := range ks {
		if err := st.RemoveDataMapping(k); err != nil {
			errors = append(errors, err)
		}
	}

	errs := []string{}
	for _, e := range errors {
		errs = append(errs, e.Error())
	}
	if len(errs) > 0 {
		log.Printf("** errors gathered: %#+v\n", strings.Join(errs, "; "))
	}
	return errors
}
