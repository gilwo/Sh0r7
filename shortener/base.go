package shortener

import (
	"github.com/eknkc/basex"
)

var (
	Base62   *basex.Encoding
	Base64SE *basex.Encoding
	Base58   *basex.Encoding
)

const (
	base62AlphaBet        = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base64SpecialAlphaBet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"
	base58AlphaBet        = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

func init() {
	var err error
	Base62, err = basex.NewEncoding(base62AlphaBet)
	if err != nil {
		panic(err)
	}
	Base64SE, err = basex.NewEncoding(base64SpecialAlphaBet)
	if err != nil {
		panic(err)
	}
	Base58, err = basex.NewEncoding(base58AlphaBet)
	if err != nil {
		panic(err)
	}
}
