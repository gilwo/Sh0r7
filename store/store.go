package store

import "time"

type Store interface {
	InitializeStore() error
	UpdateDataMapping(data []byte, short string) error
	// SaveDataMapping - store <data> under key <short> with exipry of <ttl>
	//  ttl == 0 - use default
	//  ttl < 0 dont use ttl
	SaveDataMapping(data []byte, short string, ttl time.Duration) error
	CheckExistShortDataMapping(short string) bool
	LoadDataMapping(short string) ([]byte, error)
	LoadDataMappingInfo(short string) (map[string]interface{}, error)
	SetMetaDataMapping(short, key, value string) error
	GetMetaDataMapping(short, key string) (string, error)
	RemoveDataMapping(short string) error
	GenFunc(v ...interface{}) interface{}
}

const (
	STORE_FUNC_DUMP       = "dump"
	STORE_FUNC_DUMPKEYS   = "dumpKeys"
	STORE_FUNC_GETKEYS    = "getKeys"
	STORE_FUNC_REMOVEKEYS = "removeKeys"
)

var (
	StoreCtx Store

	NewStoreLocal func() Store
	NewStoreRedis func(redisURL string) Store
)
