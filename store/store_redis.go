//go:build redis

package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	ctx           = context.Background()
	__prefixRedis string
)

type StorageRedis struct {
	redisClient *redis.Client
	redisUrl    string
	__prefix    string
}

func init() {
	NewStoreRedis = newStoreRedis
}

func newStoreRedis(redisUrl string) Store {
	return &StorageRedis{
		redisUrl: redisUrl,
		__prefix: os.Getenv("SH0R7_DEPLOY") + "$$",
	}
}

func (st *StorageRedis) _pad(short string) string {
	return st.__prefix + short
}

func (st *StorageRedis) InitializeStore() error {
	opts, err := redis.ParseURL(st.redisUrl)
	if err != nil {
		return errors.Wrapf(err, "parse url <%s> failed", st.redisUrl)
	}
	st.redisClient = redis.NewClient(opts)
	if os.Getenv("SH0R7_OTEL_UPTRACE") != "" {
		if err := redisotel.InstrumentTracing(st.redisClient); err != nil {
			return errors.Wrapf(err, "redis telemetry trace failed")
		}
		if err := redisotel.InstrumentMetrics(st.redisClient); err != nil {
			return errors.Wrapf(err, "redis telemetry metric failed")
		}
	}

	pong, err := st.redisClient.Ping(ctx).Result()
	if err != nil {
		// panic(fmt.Sprintf("Error init Redis: %v", err))
		return errors.Wrapf(err, "redis ping failed")
	}

	log.Printf("\nRedis started successfully: pong message = {%s}", pong)
	return nil
}

func (st *StorageRedis) UpdateDataMapping(data []byte, short string) error {
	entry, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err != nil {
		return errors.Wrapf(err, "redis get failed for key <%s>", short)
	}
	tup := stringTuple{}
	err = tup.unpackMsgPack([]byte(entry))
	if err != nil {
		return errors.Wrapf(err, "tuple msgunpack failed")
	}

	countNumber := 0
	sn, err := tup.AtCheck("changes")
	if err != nil {
		// never been changed
	} else {
		countNumber, err = strconv.Atoi(sn)
		if err != nil {
			return errors.New("invalid number of changes")
		}
	}
	k, err := tup.Get2(FieldDATA)
	if err != nil {
		return errors.Errorf("some problem in original data extraction")
	}
	// // keep old data
	tup.Set2(fmt.Sprintf("data_%d", countNumber), k, true)
	countNumber += 1
	tup.Set("changes", fmt.Sprintf("%d", countNumber))
	tup.Set(fmt.Sprintf("changed_%d", countNumber), time.Now().Format(time.RFC3339))
	tup.Set2Bytes(FieldDATA, data, true)

	buf, err := tup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "tuple msgpack failed")
	}
	err = st.redisClient.Set(ctx, st._pad(short), buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis set failed for <%s>", short)
	}
	return nil
}

func (st *StorageRedis) SaveDataMapping(data []byte, short string, ttl time.Duration) error {
	_, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err != redis.Nil {
		return errors.Errorf("entry exist for %s", short)
	}
	t := NewTuple()
	err = t.Set2Bytes(FieldDATA, data, true)
	if err != nil {
		return err
	}
	t.Set(FieldTime, time.Now().Format(time.RFC3339))
	if ttl == 0 {
		t.Set(FieldTTL, DefaultExpireDuration.String())
	} else if ttl > 0 {
		t.Set(FieldTTL, ttl.String())
	} // ttl < 0 - dont use ttl at all

	buf, err := t.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "tuple msgpack failed")
	}

	err = st.redisClient.Set(ctx, st._pad(short), buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis set failed for <%s>", short)
	}

	return nil
}

func (st *StorageRedis) CheckExistShortDataMapping(short string) bool {
	v, err := st.redisClient.Exists(ctx, st._pad(short)).Result()
	log.Printf("exists %s result: %v, %v\n", short, v, err)
	if err != nil || v == 1 {
		return true
	}
	return false
}
func (st *StorageRedis) LoadDataMapping(short string) ([]byte, error) {
	res, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err == redis.Nil {
		return nil, errors.Errorf("entry not exist for %s", short)
	}
	tup := stringTuple{}
	tup.unpackMsgPack([]byte(res))
	if tup.Get(FieldBlocked) == IsBLOCKED {
		return nil, errors.Errorf("not allowed %s", short)
	}
	return tup.Get2Bytes(FieldDATA)
}

func (st *StorageRedis) LoadDataMappingInfo(short string) (map[string]interface{}, error) {
	res, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err == redis.Nil {
		return nil, errors.Errorf("entry not exist for %s", short)
	}
	tup := stringTuple{}
	err = tup.unpackMsgPack([]byte(res))
	if err != nil {
		return nil, errors.Wrapf(err, "data tuple msgpack failed")
	}
	ret := map[string]interface{}{}
	for k, v := range tup.tuple {
		ret[k] = v
	}
	return ret, nil
}

func (st *StorageRedis) SetMetaDataMapping(short, key, value string) error {
	res, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err == redis.Nil {
		return errors.Errorf("entry not exist for %s", short)
	}
	tup := stringTuple{}
	err = tup.unpackMsgPack([]byte(res))
	if err != nil {
		return errors.Wrapf(err, "data tuple msgpack failed")
	}
	tup.Set(key, value)

	buf, err := tup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "tuple msg pack failed for <%s>", short)
	}
	err = st.redisClient.Set(ctx, st._pad(short), buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis set failed for <%s>", short)
	}
	return nil
}
func (st *StorageRedis) GetMetaDataMapping(short, key string) (string, error) {
	res, err := st.redisClient.Get(ctx, st._pad(short)).Result()
	if err == redis.Nil {
		return "", errors.Errorf("entry not exist for %s", short)
	}
	tup := NewTuple()
	err = tup.unpackMsgPack([]byte(res))
	if err != nil {
		return "", errors.Wrapf(err, "meta tuple msg unpack failed")
	}
	r, err := tup.AtCheck(key)
	if err != nil {
		return "", errors.Wrapf(err, "meta tuple failed for key <%s> for <%s>", key, short)
	}
	return r, nil
}
func (st *StorageRedis) RemoveDataMapping(short string) error {
	v, err := st.redisClient.Del(ctx, st._pad(short)).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.Wrapf(err, "redis key <%s> not found", short)

		}
		return errors.Wrapf(err, "redis del failed for key <%s>", short)
	}
	if v == 0 {
		return errors.Errorf("there was a problem to delete %s", short)
	}
	return nil
}

func (st *StorageRedis) GenFunc(v ...interface{}) interface{} {
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

func (st *StorageRedis) getKeys() []string {
	r, err := st.redisClient.Keys(ctx, st.__prefix+"*").Result()
	if err != nil {
		return nil
	}
	return r
}
func (st *StorageRedis) dumpKeys() string {
	r := st.getKeys()
	sort.Strings(r)
	return strings.Join(r, "\n")
}
func (st *StorageRedis) dumpAll() string {
	r := st.getKeys()
	sort.Strings(r)
	res := ""
	for _, k := range r {
		res += k + "\n" + st.dumpKey(k) + "\n\n"
	}
	return res
}
func (st *StorageRedis) dumpKey(k string) string {
	res, err := st.redisClient.Get(ctx, k).Result()
	if err == redis.Nil {
		res, err = st.redisClient.Get(ctx, st._pad(k)).Result()
		if err == redis.Nil {
			return "invalid"
		} else if err != nil {
			return "error"
		}
	}
	tup := NewTuple()
	err = tup.unpackMsgPack([]byte(res))
	if err == nil {
		return tup.Dump()
	}
	return "empty"
}

// need to test ...
func (st *StorageRedis) _removeKeys(ks []string) []error {
	log.Printf("** removing keys: %#v\n", ks)
	v, err := st.redisClient.Del(ctx, ks...).Result()
	if err != nil {
		if err == redis.Nil {
			return []error{errors.Wrapf(err, "redis key not found")}
		}
		return []error{errors.Wrapf(err, "redis del failed")}
	}
	if v == 0 {
		return []error{errors.Errorf("there was a problem to delete")}
	}
	return []error{}
}

func (st *StorageRedis) removeKeys(ks []string) []error {
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
