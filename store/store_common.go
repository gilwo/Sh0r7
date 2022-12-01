package store

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	DefaultExpireDuration = 12 * time.Hour
)

func init() {
	if expireEnv, ok := os.LookupEnv("SH0R7_EXPIRATION"); ok {
		if expire, err := time.ParseDuration(expireEnv); err != nil {
			log.Printf("failed to parse duration from env")
		} else {
			DefaultExpireDuration = expire
		}
		log.Printf("expire duration loaded from env and set to %s\n", DefaultExpireDuration)
	}
}

func Maintainence() {
	keysToDelete := []string{}
	r := StoreCtx.GenFunc(STORE_FUNC_GETKEYS)
	if r == nil {
		log.Println("unable to invoke getKeys on storage provider")
		return
	}
	keys, ok := r.([]string)
	if !ok {
		log.Println("interface assertion to []string failed")
		return
	}
	for _, k := range keys {
		info, err := StoreCtx.LoadDataMappingInfo(k)
		// short, ok := info["s"]
		if err != nil {
			log.Printf("skipping key: <%s>, err: %s\n", k, err)
			continue
		}

		v, ok := info["created"]
		if !ok {
			log.Printf("failed to get created time value for on key: <%s>\n", k)
			continue
		}
		when := v.(string)
		// if time.Parse(when)
		t, err := time.Parse(time.RFC3339, when)
		if err != nil {
			log.Printf("failed to get parse time <%s>: %s\n", when, err)
			continue
		}
		fmt.Printf("created time for key: <%s> : before parse [%s], after parse [%s]\n", k, when, t)
		v, ok = info["ttl"]
		if !ok {
			log.Printf("failed to get ttl value on key: <%s>\n", k)
			continue
		}
		ttl, err := time.ParseDuration(v.(string))
		if err != nil {
			log.Printf("failed to get parse duration <%s>: %s\n", v.(string), err)
			continue
		}

		if time.Since(t) > ttl {
			if _, ok := info["s"]; ok {
				log.Printf("all entries related to <%s> are to be deleted\n", k)
				public := info["s"].(string)
				private := info["p"].(string) + "p"
				delete := info["d"].(string) + "d"
				keysToDelete = append(keysToDelete, private, public, delete)
				if _, ok := info["url"]; ok {
					keysToDelete = append(keysToDelete, public+"url")
				}
			} else {
				keysToDelete = append(keysToDelete, k)
			}
		} else {
			log.Printf("skipping entries related to <%s> - ttl : %v, since creation %v\n", k, ttl, time.Since(t))
		}
	}
	r = StoreCtx.GenFunc(STORE_FUNC_REMOVEKEYS, keysToDelete)
	if r == nil {
		log.Println("unable to invoke removeKeys on storage provider")
		return
	}
	errors, ok := r.([]error)
	if !ok {
		log.Println("interface assertion to []error failed")
	}
	if len(errors) > 0 {
		errs := []string{}
		for _, e := range errors {
			errs = append(errs, e.Error())
		}
		log.Printf("maintainence errors gathered: %s\n", strings.Join(errs, "; "))

	}
}
