package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gilwo/Sh0r7/shortener"
	"golang.org/x/crypto/argon2"
)

var (
	bases map[string](func(v ...string) string) = map[string](func(v ...string) string){
		"58":     func(v ...string) string { return shortener.Base58.Encode([]byte(v[0])) },
		"62":     func(v ...string) string { return shortener.Base62.Encode([]byte(v[0])) },
		"64se":   func(v ...string) string { return shortener.Base64SE.Encode([]byte(v[0])) },
		"r64se":  func(v ...string) string { z, _ := shortener.Base64SE.Decode(v[0]); return string(z) },
		"arg2id": func(v ...string) string { return argon2id(v[0], v[1]) },
		// "arg2i":  func(v ...string) string { return argon2i(v[0], v[1]) },
		"arg2i":  func(v ...string) string { return argon2_(v[0], v[1], "i", "32") },
		"arg2ix": func(v ...string) string { return argon2_(v[0], v[1], "i", v[2]) },
	}
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("not enough arguments")
		return
	}

	base, ok := bases[os.Args[1]]
	if !ok {
		fmt.Printf("base %s not supported\n", os.Args[1])
		return
	}

	v := os.Args[2:]
	res := base(v...)
	fmt.Printf("%#+v: <%s>\n", v, res)

	t0 := time.Now().Truncate(time.Hour * 24) //.Round(-time.Hour * 24)
	t1 := time.Now()
	fmt.Printf("%s\n", t0)
	fmt.Printf("%s\n", t1)
	t2 := time.Since(t0).Seconds()
	fmt.Printf("%d\n", int64(t2))
}

func argon2id(pass, salt string) string {
	aVersion := argon2.Version
	aVariant := "id"
	aTime := 5
	aMemory := 4 * 1024
	aParallel := 1
	aSalt := salt
	if len(aSalt) < 8 {
		panic("salt len must be at least 8 character long")
	}
	aKeyLen := 32
	hash := argon2.IDKey([]byte(pass), []byte(aSalt), uint32(aTime), uint32(aMemory), uint8(aParallel), uint32(aKeyLen))
	encodedHash := fmt.Sprintf("$argon2%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		aVariant, aVersion,
		aMemory, aTime, aParallel,
		base64.RawStdEncoding.Strict().EncodeToString([]byte(aSalt)),
		base64.RawStdEncoding.Strict().EncodeToString([]byte(hash)))
	return encodedHash
}

func argon2i(pass, salt string) string {
	aVersion := argon2.Version
	aVariant := "i"
	aTime := 5
	aMemory := 4 * 1024
	aParallel := 1
	aSalt := salt
	if len(aSalt) < 8 {
		panic("salt len must be at least 8 character long")
	}
	aKeyLen := 32
	hash := argon2.Key([]byte(pass), []byte(aSalt), uint32(aTime), uint32(aMemory), uint8(aParallel), uint32(aKeyLen))
	encodedHash := fmt.Sprintf("$argon2%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		aVariant, aVersion,
		aMemory, aTime, aParallel,
		base64.RawStdEncoding.Strict().EncodeToString([]byte(aSalt)),
		base64.RawStdEncoding.Strict().EncodeToString([]byte(hash)))
	return encodedHash
}

func argon2_(pass, salt, variant, keylen string) string {
	aVersion := argon2.Version
	aVariant := variant
	if aVariant != "i" && aVariant != "id" {
		panic("only \"i\" and \"id\" variants are available")
	}
	var foo func(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte
	if aVariant == "i" {
		foo = argon2.Key
	} else { // "id"
		foo = argon2.IDKey
	}
	aTime := 5
	aMemory := 4 * 1024
	aParallel := 1
	aSalt := salt
	if len(aSalt) < 8 {
		panic("salt len must be at least 8 character long")
	}
	aKeyLen, err := strconv.Atoi(keylen)
	if err != nil {
		panic(err)
	}
	hash := foo([]byte(pass), []byte(aSalt), uint32(aTime), uint32(aMemory), uint8(aParallel), uint32(aKeyLen))
	encodedHash := fmt.Sprintf("$argon2%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		aVariant, aVersion,
		aMemory, aTime, aParallel,
		base64.RawStdEncoding.Strict().EncodeToString([]byte(aSalt)),
		base64.RawStdEncoding.Strict().EncodeToString([]byte(hash)))
	return encodedHash
}
