package shortener

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"github.com/gilwo/Sh0r7/store"
	"github.com/google/uuid"
	"github.com/itchyny/base58-go"
	"golang.org/x/crypto/sha3"
	"golang.org/x/crypto/twofish"
)

func GenerateToken(data string) string {
	return GenerateTokenWithStore(data, nil)
}

func GenerateTokenWithStore(data string, checkInStore store.Store) string {
	return generateShortFrom(base64.StdEncoding.EncodeToString(sha3Of(data)), -1, 0, 0, checkInStore)
}

func GenerateTokenTweaked(data string, startOffset, sizeFixed, sizeMin int) string {
	return GenerateTokenTweakedWithStore(data, startOffset, sizeFixed, sizeMin, nil)
}

func GenerateTokenTweakedWithStore(data string, startOffset, sizeFixed, sizeMin int, checkInStore store.Store) string {
	return generateShortFrom(base64.StdEncoding.EncodeToString(sha3Of(data)), startOffset, sizeFixed, sizeMin, checkInStore)
}

// generateShortFrom use input hash as a source for the short form
// startOffset - offset from the start of the input source, -1 stand for random offset
// sizeFixed - the result string length - take precedence over sizeMin, 0 stand for not used
// sizeMin - the result string minimum length - used when sizeFixed is 0, 0 stand for min any size > 1
func generateShortFrom(hash string, startOffset, sizeFixed, sizeMin int, checkInStore store.Store) string {
	N := len(hash)
	c := 0
	if sizeMin <= 0 {
		sizeMin = 1
	}
	startPos := func() int {
		if startOffset > -1 {
			return startOffset
		}
		return rand.Intn(N)
	}
	ofsCalc := func(r int) int {
		if sizeFixed < 1 {
			return rand.Intn(r)
		}
		return sizeFixed
	}
	for {
		lPos := startPos()
		ofs := ofsCalc(N - lPos)
		if ofs > sizeMin && N > lPos+ofs {
			res := hash[lPos : lPos+ofs]
			if checkInStore != nil {
				if !checkInStore.CheckExistShortDataMapping(res) {
					log.Printf("attmpets [%d], N[%d], lpos[%d], ofs[%d], short <%s>, hash<%s>\n",
						c, N, lPos, ofs, res, hash)
					return res
				}
			} else {
				return res
			}
		}
		c += 1
		if c > 1000 {
			log.Printf("attmpets [%d], data [%s], N[%d], lpos[%d], ofs[%d], short <%s>, hash<%s>\n",
				c, hash, N, lPos, ofs, "no token", hash)
			log.Printf("not found proper short after %d attempts\n", c)
			return ""
		}
	}
}

func GenerateShortDataWithStore_(data string, checkInStore store.Store) string {
	urlHashBytes := sha256Of(data + uuid.NewString())
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	return generateShortFrom(base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber))), -1, 0, 0, checkInStore)
}
func GenerateShortData(data string) string {
	return GenerateShortDataWithStore(data, nil)
}

func GenerateShortDataTweakedWithStore_(data string, startOffset, sizeFixed, sizeMin int, checkInStore store.Store) string {
	urlHashBytes := sha256Of(data + uuid.NewString())
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	return generateShortFrom(base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber))), startOffset, sizeFixed, sizeMin, checkInStore)
}

func GenerateShortDataTweaked(data string, startOffset, sizeFixed, sizeMin int) string {
	return GenerateShortDataTweakedWithStore(data, startOffset, sizeFixed, sizeMin, nil)
}

func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}
func sha3Of(input string) []byte {
	h := make([]byte, 64)
	sha3.ShakeSum256(h, []byte(input))
	return h
}

func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		// fmt.Println(err.Error())
		panic(err)
	}
	return string(encoded)
}

func GenerateShortDataWithStore(data string, checkInStore store.Store) string {
	return generateShortFrom(Base58.Encode(sha3Of(data+uuid.NewString())), -1, 0, 0, checkInStore)
}
func GenerateShortDataTweakedWithStore(data string, startOffset, sizeFixed, sizeMin int, checkInStore store.Store) string {
	return generateShortFrom(Base58.Encode(sha3Of(data+uuid.NewString())), startOffset, sizeFixed, sizeMin, checkInStore)
}
func GenerateShortDataTweakedWithStore2(data string, startOffset, sizeFixed, sizeMin, sizeMax int, checkInStore store.Store) string {
	return GenericShort(Base58.Encode(sha3Of(data+uuid.NewString())), startOffset, sizeFixed, sizeMin, sizeMax, checkInStore)
}
func GenerateShortDataTweakedWithStore2NotRandom(data string, startOffset, sizeFixed, sizeMin, sizeMax int, checkInStore store.Store) string {
	return GenericShort(Base58.Encode(sha3Of(data)), startOffset, sizeFixed, sizeMin, sizeMax, checkInStore)
}

// GenericShort use original as input (which can be long) as a base for short string
// startOffset - offset from the start of the input source, -1 stand for random offset
// sizeFixed - the result short string length - take precedence over sizeMin, 0 stand for not used
// sizeMin - the result short string minimum length - used when sizeFixed is 0, 0 stand for no limit on minimum where (at least 1)
// sizeMax - the result short string maximum length - used when sizeFixed is 0, 0 stand for no limit on maximum
func GenericShort(original string, startOffset, sizeFixed, sizeMin, sizeMax int, checkInStore store.Store) string {
	N := len(original)
	c := 0
	if sizeMin <= 0 {
		sizeMin = 1
	}
	if sizeMax <= 0 {
		sizeMax = N
	}
	startPos := func() int {
		if startOffset > -1 {
			return startOffset
		}
		return rand.Intn(N)
	}
	ofsCalc := func(r int) int {
		if sizeFixed == 0 {
			return rand.Intn(func() int {
				if r >= sizeMax {
					return sizeMax
				}
				return r
			}() + 1) // half open - so we need to add +1 to the half closed side to include that
		}
		if sizeFixed < 1 {
			return rand.Intn(r)
		}
		return sizeFixed
	}
	for {
		lPos := startPos()
		ofs := ofsCalc(N - lPos)
		if ofs >= sizeMin && N > lPos+ofs {
			res := original[lPos : lPos+ofs]
			if checkInStore != nil {
				if !checkInStore.CheckExistShortDataMapping(res) {
					if c > 1 {
						log.Printf("attmpets [%d], N[%d], lpos[%d], ofs[%d], range [%d-%d], short <%s>, hash<%s>\n",
							c, N, lPos, ofs, sizeMin, sizeMax, res, original)
					}
					return res
				}
			} else {
				return res
			}
		}
		c += 1
		if c > 1000 {
			log.Printf("not found suitable short after [%d] attmpets , original [%s], N[%d], lpos[%d], ofs[%d], range [%d-%d]\n",
				c, original, N, lPos, ofs, sizeMin, sizeMax)
			return ""
		}
	}
}

func EncryptData(data []byte, key string) []byte {
	if len(data) == 0 {
		return nil
	}
	keyb := sha3Of(key)
	cipherKey, iv, pad := keyb[:32], keyb[32:48], keyb[48:56]
	log.Printf("key<%s>\n\t- cipherKey: <%s>\n\t- iv: <%s>\n\t- pad: <%s>\n",
		key,
		hex.EncodeToString(cipherKey),
		hex.EncodeToString(iv),
		hex.EncodeToString(pad))
	twofishC, err := twofish.NewCipher(cipherKey)
	if err != nil {
		log.Printf("problem with twofish cipher creation : %v\n", err)
		return nil
	}
	src := data
	cbc := cipher.NewCBCEncrypter(twofishC, iv)
	extraNeeded := cbc.BlockSize()
	if rem := len(data) % cbc.BlockSize(); rem > 0 {
		extraNeeded += cbc.BlockSize() - rem
	}
	pad = append(pad, byte(extraNeeded))
	src = append(pad, append(make([]byte, extraNeeded-len(pad)), data...)...)
	dst := make([]byte, len(src))
	_dst := dst
	_src := src
	for len(_src) > 0 {
		cbc.CryptBlocks(_dst[:cbc.BlockSize()], _src[:cbc.BlockSize()])
		_dst = _dst[cbc.BlockSize():]
		_src = _src[cbc.BlockSize():]
	}
	return dst
}

func DecryptData(data []byte, key string) []byte {
	if len(data) == 0 {
		return nil
	}
	keyb := sha3Of(key)
	cipherKey, iv, pad := keyb[:32], keyb[32:48], keyb[48:56]
	log.Printf("key<%s>\n\t- cipherKey: <%s>\n\t- iv: <%s>\n\t- pad: <%s>\n",
		key,
		hex.EncodeToString(cipherKey),
		hex.EncodeToString(iv),
		hex.EncodeToString(pad))
	twofishC, err := twofish.NewCipher(cipherKey)
	if err != nil {
		log.Printf("problem with twofish cipher creation : %v\n", err)
		return nil
	}
	src := data
	cbc := cipher.NewCBCDecrypter(twofishC, iv)

	dst := make([]byte, len(data))
	_dst := dst
	_src := src
	for len(_src) > 0 {
		cbc.CryptBlocks(_dst[:cbc.BlockSize()], _src[:cbc.BlockSize()])
		_dst = _dst[cbc.BlockSize():]
		_src = _src[cbc.BlockSize():]
	}
	if !bytes.HasPrefix(dst, pad) {
		log.Printf("decryped data missing pad")
		return nil
	}
	leftOver := bytes.TrimPrefix(dst, pad)
	padLength := uint8(leftOver[0])
	dst = dst[padLength:]
	return dst
}
