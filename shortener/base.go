package shortener

import (
	"github.com/eknkc/basex"
)

var (
	Base58 *basex.Encoding
)

const (
	base58AlphaBet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

func init() {
	var err error
	Base58, err = basex.NewEncoding(base58AlphaBet)
	if err != nil {
		panic(err)
	}
}
