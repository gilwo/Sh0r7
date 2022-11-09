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
)

func GenerateToken(data string, size int) string {
	urlHashBytes := sha256Of(data)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	urlHashBytes = sha256Of(uuid.NewString())
	generatedNumber = new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString += base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	finalString = base64.StdEncoding.EncodeToString([]byte(finalString))
	N := len(finalString)
	c := 0
	for {
		lPos := rand.Intn(N)
		ofs := size
		if N > lPos+ofs {
			fmt.Printf("attmpets [%d], data [%s], N[%d], lpos[%d], ofs[%d], short <%s>, final<%s>\n",
				c, data, N, lPos, ofs, "no token", finalString)
			fmt.Printf("not found proper short after %d attempts\n", c)
			return finalString[lPos : lPos+ofs]
		}
		c += 1
		if c > 100 {
			fmt.Printf("attmpets [%d], data [%s], N[%d], lpos[%d], ofs[%d], short <%s>, final<%s>\n",
				c, data, N, lPos, ofs, "no token", finalString)
			fmt.Printf("not found proper short after %d attempts\n", c)
			return ""
		}
	}
}
func GenerateShortData(data string) string {
	urlHashBytes := sha256Of(data + uuid.NewString())
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))

	N := len(finalString)
	c := 0
	for {
		lPos := rand.Intn(N)
		ofs := rand.Intn(10)
		if ofs > 1 && N > lPos+ofs {
			res := finalString[lPos : lPos+ofs]
			if store.StoreCtx.CheckShortDataMapping(res) == nil {
				if c > 1 {
					fmt.Printf("found proper short after %d attempts\n", c)
				}
				fmt.Printf("attmpets [%d], N[%d], lpos[%d], ofs[%d], short <%s>, final<%s>\n",
					c, N, lPos, ofs, res, finalString)
				return res
			}
		}
		c += 1
		if c > 100 {
			fmt.Printf("not found proper short after %d attempts\n", c)
			return ""
		}
	}
}
func GenerateShortLink(initialLink string, userId string) string {
	urlHashBytes := sha256Of(initialLink + userId)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:8]
}

func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
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
