package store

type Store interface {
	InitializeStore() error
	UpdateDataMapping(data []byte, short string) error
	SaveDataMapping(data []byte, short string) error
	CheckShortDataMapping(short string) error
	LoadDataMapping(short string) ([]byte, error)
	LoadDataMappingInfo(short string) (map[string]interface{}, error)
	SetMetaDataMapping(short, key, value string) error
	GetMetaDataMapping(short, key string) (string, error)
	RemoveDataMapping(short string) error
	GenFunc(v ...interface{}) interface{}
}

var (
	StoreCtx Store

	NewStoreLocal func() Store
	NewStoreRedis func(redisURL string) Store
)

/*
NewStringTuple(values ...*fieldValue) (*stringTuple, error) {
(t *stringTuple) AtCheck(field string) (string, error) {
(t *stringTuple) Get(field string) string {
(t *stringTuple) Set(field, value string) {
(t *stringTuple) SetCheck(field, value string) error {
(t *stringTuple) Keys() []string {
*/
