package shortener

import (
	"bytes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strings"

	"github.com/gilwo/Sh0r7/store"
	"github.com/google/uuid"
	"github.com/itchyny/base58-go"
	"golang.org/x/crypto/argon2"
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
	// log.Printf("&&& sha30 input: <%s>\n", input)
	h := make([]byte, 64)
	sha3.ShakeSum256(h, []byte(input))
	// log.Printf("&&& sha30 res: <%#v>\n", h)
	// log.Printf("&&& sha30 encode to string: <%s>\n", base64.StdEncoding.EncodeToString(h))
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
		// fmt.Printf("sizefixed: %d, r: %d\n", sizeFixed, r)
		if sizeFixed == 0 {
			return rand.Intn(func() int {
				// if r <= sizeMax {
				// 	return r
				// }
				if r >= sizeMax {
					// fmt.Printf("r>sizemiax: (%d>%d)\n", r, sizeMax)
					return sizeMax
				}
				// fmt.Printf("r<=sizemiax: (%d>%d)\n", r, sizeMax)
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
		// fmt.Printf("ofsCalc: %d\n", ofs)
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

var (
	Env   string = "dev"
	isDev bool   = __isDev(Env)
)

func __isDev(in string) bool {
	if in == "dev" {
		return true
	}
	return false
}

// func __Vargon2Generic(variant, pass, salt, keylen string) (string, error) {
// 	aVersion := argon2.Version
// 	aVariant := variant
// 	if aVariant != "i" && aVariant != "id" {
// 		msg := "only \"i\" and \"id\" variants are available"
// 		if isDev {
// 			panic(msg)
// 		}
// 		return "", fmt.Errorf(msg)
// 	}
// 	var foo func(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte
// 	if aVariant == "i" {
// 		foo = argon2.Key
// 	} else { // "id"
// 		foo = argon2.IDKey
// 	}
// 	aTime := 5
// 	aMemory := 4 * 1024
// 	aParallel := 1
// 	aSalt := salt
// 	if len(aSalt) < 8 {
// 		msg := "salt len must be at least 8 character long"
// 		if isDev {
// 			panic(msg)
// 		}
// 		return "", fmt.Errorf(msg)
// 	}
// 	aKeyLen, err := strconv.Atoi(keylen)
// 	if err != nil {
// 		if isDev {
// 			panic(err)
// 		}
// 		return "", err
// 	}
// 	hash := foo([]byte(pass), []byte(aSalt), uint32(aTime), uint32(aMemory), uint8(aParallel), uint32(aKeyLen))
// 	encodedHash := fmt.Sprintf("$argon2%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
// 		aVariant, aVersion,
// 		aMemory, aTime, aParallel,
// 		base64.RawStdEncoding.Strict().EncodeToString([]byte(aSalt)),
// 		base64.RawStdEncoding.Strict().EncodeToString([]byte(hash)))
// 	return encodedHash, nil
// }

func __argon2Generic(variant, pass, salt string, keylen uint32) (string, error) {
	aVersion := argon2.Version
	aVariant := variant
	if aVariant != "i" && aVariant != "id" {
		msg := "only \"i\" and \"id\" variants are available"
		if isDev {
			panic(msg)
		}
		return "", fmt.Errorf(msg)
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
		msg := "salt len must be at least 8 character long"
		if isDev {
			panic(msg)
		}
		return "", fmt.Errorf(msg)
	}
	aKeyLen := keylen
	// aKeyLen, err := strconv.Atoi(keylen)
	// if err != nil {
	// 	if isDev {
	// 		panic(err)
	// 	}
	// 	return "", err
	// }
	hash := foo([]byte(pass), []byte(aSalt), uint32(aTime), uint32(aMemory), uint8(aParallel), uint32(aKeyLen))
	encodedHash := fmt.Sprintf("$argon2%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		aVariant, aVersion,
		aMemory, aTime, aParallel,
		base64.RawStdEncoding.Strict().EncodeToString([]byte(aSalt)),
		base64.RawStdEncoding.Strict().EncodeToString([]byte(hash)))
	return encodedHash, nil
}

// Argon2I - preferred for hasing passwords
func Argon2I(pass, salt string, keylen uint32) string {
	res, _ := __argon2Generic("i", pass, salt, keylen)
	return res
}

func Argon2ID(pass, salt string, keylen uint32) string {
	res, _ := __argon2Generic("id", pass, salt, keylen)
	return res
}

func DecodeArgonHashSalt(encodedHash string) string {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return ""
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return ""
	}
	return string(salt)

}

//ref ...
// func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
// 	vals := strings.Split(encodedHash, "$")
// 	if len(vals) != 6 {
// 		return nil, nil, nil, ErrInvalidHash
// 	}

// 	var version int
// 	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	if version != argon2.Version {
// 		return nil, nil, nil, ErrIncompatibleVersion
// 	}

// 	p = &params{}
// 	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}

// 	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	p.saltLength = uint32(len(salt))

// 	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	p.keyLength = uint32(len(hash))

//		return p, salt, hash, nil
//	}

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
	if len(data) < 2*cbc.BlockSize() {
		return nil
	}

	dst := make([]byte, len(data))
	_dst := dst
	_src := src
	log.Printf("src len :%d\n", len(_src))
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

func CreateECKeyPairString() (string, string, error) {
	if kPair, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader); err != nil {
		return "", "", err
	} else if prvEncoded, err := x509.MarshalECPrivateKey(kPair); err != nil {
		return "", "", err
	} else if pubEncoded, err := x509.MarshalPKIXPublicKey(&kPair.PublicKey); err != nil {
		return "", "", err
	} else {
		return hex.EncodeToString(prvEncoded), hex.EncodeToString(pubEncoded), nil
	}
}

func TransformECKeyPrivateString(prv string) *ecdsa.PrivateKey {
	if prvData, err := hex.DecodeString(prv); err != nil {
		return nil
	} else if prvKey, err := x509.ParseECPrivateKey(prvData); err != nil {
		return nil
	} else {
		return prvKey
	}
}

func TransformECKeyPublicString(pub string) *ecdsa.PublicKey {
	if pubData, err := hex.DecodeString(pub); err != nil {
		return nil
	} else if genPubKey, err := x509.ParsePKIXPublicKey(pubData); err != nil {
		return nil
	} else {
		return genPubKey.(*ecdsa.PublicKey)
	}
}

func TransformECKeyPairString(prv, pub string) *ecdsa.PrivateKey {

	if prvData, err := hex.DecodeString(prv); err != nil {
		return nil
	} else if pubData, err := hex.DecodeString(pub); err != nil {
		return nil
	} else if prvKey, err := x509.ParseECPrivateKey(prvData); err != nil {
		return nil
	} else if genPubKey, err := x509.ParsePKIXPublicKey(pubData); err != nil {
		return nil
	} else {
		pubKey := genPubKey.(*ecdsa.PublicKey)
		prvKey.PublicKey = *pubKey
		return prvKey
	}
}

func signASN(prv, in string) string {
	prvK := TransformECKeyPrivateString(prv)
	if prvK == nil {
		return ""
	}
	hash := sha3Of(in)
	signatureASN, err := ecdsa.SignASN1(crand.Reader, prvK, hash)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(signatureASN)
}

func verifyASN(pub, in, signature string) bool {
	pubK := TransformECKeyPublicString(pub)
	if pubK == nil {
		return false
	}
	signData, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := sha3Of(in)
	return ecdsa.VerifyASN1(pubK, hash, signData)
}

func sign(prv, pub, in string) string {
	k := TransformECKeyPairString(prv, pub)
	if k == nil {
		return ""
	}
	hash := sha3Of(in)
	r, s, err := ecdsa.Sign(crand.Reader, k, hash)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(r.Bytes()) + hex.EncodeToString(hash) + hex.EncodeToString(s.Bytes())

	// signature := append(r.Bytes(), s.Bytes()...)
	// return hex.EncodeToString(signature)
}

func verify(prv, pub, in, signature string) bool {
	k := TransformECKeyPairString(prv, pub)
	if k == nil {
		return false
	}
	hash := sha3Of(in)

	z := strings.Split(signature, hex.EncodeToString(hash))
	r := &big.Int{}
	s := &big.Int{}

	rin, err := hex.DecodeString(z[0])
	if err != nil {
		return false
	}
	sin, err := hex.DecodeString(z[1])
	if err != nil {
		return false
	}

	return ecdsa.Verify(&k.PublicKey, hash, r.SetBytes(rin), s.SetBytes(sin))
}

func main2() {
	prv, pub, err := CreateECKeyPairString()
	if err != nil {
		panic(err)
	}
	shoobi := "shoobi doobi"
	signature := sign(prv, pub, shoobi)

	check := verify(prv, pub, shoobi, signature)

	// kp := TransformECKeyPairString(prv, pub)
	// if kp == nil {
	// 	panic("invalid keypair")
	// }
	fmt.Printf("check sign/verify: %v\n", check)
}

func main3() {
	prv, pub, err := CreateECKeyPairString()
	if err != nil {
		panic(err)
	}
	shoobi := "shoobi doobi"
	signature := signASN(prv, shoobi)

	check := verifyASN(pub, shoobi, signature)

	// kp := TransformECKeyPairString(prv, pub)
	// if kp == nil {
	// 	panic("invalid keypair")
	// }
	fmt.Printf("check sign/verify: %v\n", check)
}

/// ---------------------------------------------------
