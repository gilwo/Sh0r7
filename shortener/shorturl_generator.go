package shortener

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/gilwo/Sh0r7/store"
	"github.com/google/uuid"
	"github.com/itchyny/base58-go"
	"golang.org/x/crypto/sha3"
)

func GenerateToken(data string) string {
	return generateShortFrom(base64.StdEncoding.EncodeToString(sha3Of(data)), -1, 0, 0)
}

func GenerateTokenTweaked(data string, startOffset, sizeFixed, sizeMin int) string {
	return generateShortFrom(base64.StdEncoding.EncodeToString(sha3Of(data)), startOffset, sizeFixed, sizeMin)
}

// generateShortFrom use input hash as a source for the short form
// startOffset - offset from the start of the input source, -1 stand for random offset
// sizeFixed - the result string length - take precedence over sizeMin, 0 stand for not used
// sizeMin - the result string minimum length - used when sizeFixed is 0, 0 stand for min any size > 1
func generateShortFrom(hash string, startOffset, sizeFixed, sizeMin int) string {
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
			if store.StoreCtx.CheckShortDataMapping(res) == nil {
				fmt.Printf("attmpets [%d], N[%d], lpos[%d], ofs[%d], short <%s>, hash<%s>\n",
					c, N, lPos, ofs, res, hash)
				return res
			}
		}
		c += 1
		if c > 1000 {
			fmt.Printf("attmpets [%d], data [%s], N[%d], lpos[%d], ofs[%d], short <%s>, hash<%s>\n",
				c, hash, N, lPos, ofs, "no token", hash)
			fmt.Printf("not found proper short after %d attempts\n", c)
			return ""
		}
	}
}

func GenerateShortData(data string) string {
	urlHashBytes := sha256Of(data + uuid.NewString())
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	return generateShortFrom(base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber))), -1, 0, 0)
}

func GenerateShortDataTweaked(data string, startOffset, sizeFixed, sizeMin int) string {
	urlHashBytes := sha256Of(data + uuid.NewString())
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	return generateShortFrom(base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber))), startOffset, sizeFixed, sizeMin)
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
