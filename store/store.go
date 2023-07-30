package store

import "time"

type Store interface {
	// FIXME - is it needed?
	InitializeStore() error
	// Update the data associated with the short -
	//
	// this will keep history (old data/number of changes/change time) for the short and
	UpdateDataMapping(data []byte, short string) error
	// SaveDataMapping - store <data> under key <short> with exipry of <ttl>
	//  ttl == 0 - use default
	//  ttl < 0 dont use ttl
	SaveDataMapping(data []byte, short string, ttl time.Duration) error
	// Kinda obvious, is this short exists, lock the short so no one can use it if its not exist
	CheckExistShortDataMappingAndLock(short string) bool
	// Kinda obvious, is this short exists
	CheckExistShortDataMapping(short string) bool
	// Retrieve only the data associated with the short
	LoadDataMapping(short string) ([]byte, error)
	// Retrieve data and metadata associated with the short
	LoadDataMappingInfo(short string) (map[string]interface{}, error)
	// Associate metadata with the short in the form of key/value
	//  Note - value is overwritten
	SetMetaDataMapping(short, key, value string) error
	// Retrieve metadata associated with the short according to key
	GetMetaDataMapping(short, key string) (string, error)
	// Remove the short and all associated info
	RemoveDataMapping(short string) error
	// General purpose func
	GenFunc(v ...interface{}) interface{}
}

const (
	STORE_FUNC_DUMP       = "dump"
	STORE_FUNC_DUMPKEYS   = "dumpKeys"
	STORE_FUNC_DUMPALL    = "dumpAll"
	STORE_FUNC_GETKEYS    = "getKeys"
	STORE_FUNC_REMOVEKEYS = "removeKeys"
)

var (
	// handler to storage provider
	StoreCtx Store
	// local storage provider initialization
	NewStoreLocal func() Store
	// redis storage provider initialization
	NewStoreRedis func(redisURL string) Store
)
