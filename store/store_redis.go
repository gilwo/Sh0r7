// go:build redis

package store

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	ctx = context.Background()
)

const CacheDuration = 6 * time.Hour

type StorageRedis struct {
	redisClient *redis.Client
	redisUrl    string
}

func init() {
	NewStoreRedis = newStoreRedis
}

func newStoreRedis(redisUrl string) Store {
	return &StorageRedis{
		redisUrl: redisUrl,
	}
}
func (st *StorageRedis) InitializeStore() error {
	opts, err := redis.ParseURL(st.redisUrl)
	if err != nil {
		return errors.Wrapf(err, "parse url <%s> failed", st.redisUrl)
	}
	st.redisClient = redis.NewClient(opts)

	pong, err := st.redisClient.Ping(ctx).Result()
	if err != nil {
		// panic(fmt.Sprintf("Error init Redis: %v", err))
		return errors.Wrapf(err, "redis ping failed")
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	return nil
}

func (st *StorageRedis) UpdateDataMapping(data []byte, short string) error {
	prevData, err := st.redisClient.Get(ctx, short).Result()
	if err != nil {
		return errors.Wrapf(err, "redis get failed for key <%s>", short)
	}
	prevDataTup := stringTuple{}
	err = prevDataTup.unpackMsgPack([]byte(prevData))
	if err != nil {
		return errors.Wrapf(err, "tuple msgunpack failed")
	}

	s := base64.StdEncoding.EncodeToString(data)
	var in bytes.Buffer
	b := []byte(s)
	w, err := zlib.NewWriterLevel(&in, zlib.BestCompression)
	if err != nil {
		return errors.Wrapf(err, "zlib writer failed")
	}
	w.Write(b)
	w.Close()

	k := base64.StdEncoding.EncodeToString(in.Bytes())
	tup, err := NewStringTuple([]*fieldValue{{"data", k}, {"changed", time.Now().String()}, {"created", prevDataTup.Get("created")}}...)
	if err != nil {
		return errors.Wrapf(err, "new string tuple failed")
	}

	buf, err := tup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "tuple msgpack failed")
	}

	err = st.redisClient.Set(ctx, short, buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis set failed for <%s>", short)
	}

	metaTup := NewTuple()
	meta, err := st.redisClient.Get(ctx, short+".meta").Result()
	if err != nil {
		if err != redis.Nil {
			return errors.Wrapf(err, "redis meta get failed for <%s>", short)
		}
		return errors.Wrapf(err, "redis meta get not exists for <%s>", short)
	}
	err = metaTup.unpackMsgPack([]byte(meta))
	if err != nil {
		return errors.Wrapf(err, "meta tuple msg unpack failed for <%s>", short)
	}
	fmt.Printf("meta tuple:\n%s\n", metaTup.Dump())

	countNumber := 0
	sn, err := metaTup.AtCheck("changes")
	if err == nil {
		countNumber, err = strconv.Atoi(sn)
		if err != nil {
			return errors.New("invalid number of changes")
		}
	}
	// // keep old data
	metaTup.Set(fmt.Sprintf("data_%d", countNumber), prevDataTup.Get("data"))
	countNumber += 1
	metaTup.Set("changes", fmt.Sprintf("%d", countNumber))
	metaTup.Set(fmt.Sprintf("changed_time_%d", countNumber), time.Now().String())

	buf2, err := metaTup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "meta tuple msgpack failed")
	}
	err = st.redisClient.Set(ctx, short+".meta", buf2, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis meta set failed for <%s>", short)
	}
	return nil
}

func (st *StorageRedis) SaveDataMapping(data []byte, short string) error {
	_, err := st.redisClient.Get(ctx, short).Result()
	if err != redis.Nil {
		return errors.Errorf("entry exist for %s", short)
	}
	tup := NewTuple()
	err = tup.Set2Bytes("data", data, true)
	if err != nil {
		return err
	}
	tup.Set("created", time.Now().String())

	buf, err := tup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "tuple msgpack failed")
	}

	err = st.redisClient.Set(ctx, short, buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis set failed for <%s>", short)
	}

	return nil
}

func (st *StorageRedis) CheckShortDataMapping(short string) error {
	v, err := st.redisClient.Exists(ctx, short).Result()
	fmt.Printf("exists %s result: %v, %v\n", short, v, err)
	if err != nil || v == 1 {
		return errors.Errorf("entry exist for %s", short)
	}
	return nil
}
func (st *StorageRedis) LoadDataMapping(short string) ([]byte, error) {
	res, err := st.redisClient.Get(ctx, short).Result()
	if err == redis.Nil {
		return nil, errors.Errorf("entry not exist for %s", short)
	}
	tup := stringTuple{}
	tup.unpackMsgPack([]byte(res))
	return tup.Get2Bytes("data")
}

func (st *StorageRedis) LoadDataMappingInfo(short string) (map[string]interface{}, error) {
	res, err := st.redisClient.Get(ctx, short).Result()
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
	meta, err := st.redisClient.Get(ctx, short+".meta").Result()
	if err != nil {
		return nil, errors.Wrapf(err, "redis meta get failed for <%s>", short)
	}
	metaTup := NewTuple()
	err = metaTup.unpackMsgPack([]byte(meta))
	if err != nil {
		return nil, errors.Wrapf(err, "meta tuple msg unpack failed")
	}
	for _, k := range metaTup.Keys() {
		ret[k] = metaTup.Get(k)
	}
	return ret, nil
}

func (st *StorageRedis) SetMetaDataMapping(short, key, value string) error {
	_, err := st.redisClient.Keys(ctx, short).Result()
	if err == redis.Nil {
		return errors.Errorf("entry not exist for %s", short)
	}
	metaTup := NewTuple()
	meta, err := st.redisClient.Get(ctx, short+".meta").Result()
	if err != nil {
		if err != redis.Nil {
			return errors.Wrapf(err, "redis meta get failed for <%s>", short)
		}
	} else {
		err = metaTup.unpackMsgPack([]byte(meta))
		if err != nil {
			return errors.Wrapf(err, "meta tuple msg unpack failed for <%s>", short)
		}
	}
	metaTup.Set(key, value)
	buf, err := metaTup.packMsgPack()
	if err != nil {
		return errors.Wrapf(err, "meta tuple msg pack failed for <%s>", short)
	}
	err = st.redisClient.Set(ctx, short+".meta", buf, 0).Err()
	if err != nil {
		return errors.Wrapf(err, "redis meta set failed for <%s>", short)
	}
	return nil
}
func (st *StorageRedis) GetMetaDataMapping(short, key string) (string, error) {
	_, err := st.redisClient.Get(ctx, short).Result()
	if err == redis.Nil {
		return "", errors.Errorf("entry not exist for %s", short)
	}
	v, err := st.redisClient.Get(ctx, short+".meta").Result()
	if err != nil {
		return "", errors.Wrapf(err, "redis get meta failed for key <%s>", short)
	}
	tup := NewTuple()
	err = tup.unpackMsgPack([]byte(v))
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
	_, err := st.redisClient.Get(ctx, short).Result()
	if err == redis.Nil {
		return errors.Errorf("entry not exist for %s", short)
	}
	v, err := st.redisClient.Del(ctx, short, short+".meta").Result()
	if err != nil {
		return errors.Wrapf(err, "redis del failed for key <%s>", short)
	}
	if v == 0 {
		return errors.Errorf("there was a problem to delete %s", short)
	}
	return nil
}
