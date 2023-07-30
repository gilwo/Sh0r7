package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	ecies "github.com/ecies/go/v2"
	"golang.org/x/crypto/sha3"
)

func sha3Of(input string) []byte {
	h := make([]byte, 64)
	sha3.ShakeSum256(h, []byte(input))
	return h
}

func CreateECKeyPairString() (string, string, error) {
	if kPair, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader); err != nil {
		return "", "", err
	} else if prvEncoded, err := x509.MarshalECPrivateKey(kPair); err != nil {
		return "", "", err
	} else if pubEncoded, err := x509.MarshalPKIXPublicKey(&kPair.PublicKey); err != nil {
		return "", "", err
	} else {
		log.Printf("ecdsa\nprv: %#+v\npub: %#+v\n", kPair, kPair.PublicKey)
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
	fmt.Printf("check sign/verify: %v\n", check)
}

func main3() {
	defer func() {
		x := recover()
		if x != nil {
			log.Printf("!! panic recovered with :%T, %v", x, x)
			// this would happen because the ecdsa key generated based on p384 curve
			// while the ecies key expect a secp256k1 curve for ecies encryption logic
		}
	}()

	prv, pub, err := CreateECKeyPairString()
	if err != nil {
		panic(err)
	}
	shoobi := "shoobi doobi"
	signature := signASN(prv, shoobi)

	check := verifyASN(pub, shoobi, signature)
	fmt.Printf("check ASN sign/verify: %v\n", check)

	cipher, err := fooEnc(prv, pub, shoobi)
	if err != nil {
		fmt.Printf("encryption failed: %s\n", err)
		return
	}
	txt, err := fooDec(prv, pub, cipher)
	if err != nil {
		fmt.Printf("decryption failed: %s\n", err)
		return
	}
	fmt.Printf("txt: [%s]\n", txt)
}
func main4() {
	prv, pub, err := CreateECKeyPairString_ecies()
	if err != nil {
		panic(err)
	}
	log.Printf("\n--prv: %s\n", prv)
	log.Printf("\n--pub: %s\n", pub)
	shoobi := "shoobi doobi"
	// signature := signASN(prv, shoobi)

	// check := verifyASN(pub, shoobi, signature)
	// fmt.Printf("check ASN sign/verify: %v\n", check)

	cipher, err := fooEnc_ecies(prv, pub, shoobi)
	if err != nil {
		fmt.Printf("encryption failed: %s\n", err)
		return
	}
	txt, err := fooDec_ecies(prv, pub, cipher)
	if err != nil {
		fmt.Printf("decryption failed: %s\n", err)
		return
	}
	fmt.Printf("txt: [%s]\n", txt)

	signature := sign_ecies(prv, pub, shoobi)
	check := verify_ecies(pub, shoobi, signature)
	fmt.Printf("check ecies sign/verify: %v\n", check)
}

func main() {
	main2()
	main3()
	main4()
}

func fooEnc(prv, pub, in string) (string, error) {
	// prvKeyByte, err := hex.DecodeString(prv)
	// if err != nil {
	// 	return "", err
	// }
	// pubKeyByte, err := hex.DecodeString(pub)
	// if err != nil {
	// 	return "", err
	// }
	// prvKey := ecies.NewPrivateKeyFromBytes(prvKeyByte)
	// prvKey.PublicKey, err = ecies.NewPublicKeyFromBytes(pubKeyByte)
	// if err != nil {
	// 	return "", err
	// }
	prvKey, err := _convertKeys(prv, pub)
	if err != nil {
		return "", err
	}

	log.Printf("ecies\nprv: %#+v\npub: %#+v\n", prvKey, prvKey.PublicKey)
	outByte, err := ecies.Encrypt(prvKey.PublicKey, []byte(in))
	return hex.EncodeToString(outByte), err
}

func fooDec(prv, pub, in string) (string, error) {
	// key := TransformECKeyPairString(prv, pub)
	// if key == nil {
	// 	return "", fmt.Errorf("invalid prv/pub keypair")
	// }

	// prvKeyByte, err := hex.DecodeString(prv)
	// if err != nil {
	// return "", err
	// }
	// pubKeyByte, err := hex.DecodeString(pub)
	// if err != nil {
	// return "", err
	// }
	// prvKey := ecies.NewPrivateKeyFromBytes(prvKeyByte)
	// prvKey.PublicKey, err = ecies.NewPublicKeyFromBytes(pubKeyByte)
	// if err != nil {
	// return "", err
	// }

	// // pryKeu &:= ecies.PrivateKey{
	// 	PublicKey: &ecies.PublicKey{
	// 		Curve: key.PublicKey.Curve,
	// 		X:     big.NewInt(key.PublicKey.X.Int64()),
	// 		Y:     big.NewInt(key.PublicKey.Y.Int64()),
	// 	},
	// 	D: big.NewInt(key.D.Int64()),
	// }
	inByte, err := hex.DecodeString(in)
	if err != nil {
		return "", err
	}
	prvKey, err := _convertKeys(prv, pub)
	if err != nil {
		return "", err
	}
	log.Printf("ecies\nprv: %#+v\npub: %#+v\n", prvKey, prvKey.PublicKey)
	out, err := ecies.Decrypt(prvKey, inByte)
	return string(out), err
}

func _convertKeys(prv, pub string) (*ecies.PrivateKey, error) {
	key := TransformECKeyPairString(prv, pub)
	if key == nil {
		return nil, fmt.Errorf("invalid prv/pub keypair")
	}
	return &ecies.PrivateKey{
		D: key.D,
		PublicKey: &ecies.PublicKey{
			Curve: key.PublicKey.Curve,
			X:     key.PublicKey.X,
			Y:     key.PublicKey.Y,
		},
	}, nil
}

func CreateECKeyPairString_ecies() (string, string, error) {
	if kPair, err := ecies.GenerateKey(); err != nil {
		return "", "", err
	} else {
		log.Printf("ecies\nprv: %#+v\npub: %#+v\n", kPair, kPair.PublicKey)
		return kPair.Hex(), kPair.PublicKey.Hex(true), nil
	}
	// } else if prvEncoded, err := kPair.; err != nil {
	// 	return "", "", err
	// } else if pubEncoded, err := x509.MarshalPKIXPublicKey(&kPair.PublicKey); err != nil {
	// 	return "", "", err
	// } else {
	// 	return hex.EncodeToString(prvEncoded), hex.EncodeToString(pubEncoded), nil
	// }
}

func fooEnc_ecies(prv, pub, in string) (string, error) {
	keyPub, err := ecies.NewPublicKeyFromHex(pub)
	if err != nil {
		return "", err
	}
	outByte, err := ecies.Encrypt(keyPub, []byte(in))
	return hex.EncodeToString(outByte), err
}

func fooDec_ecies(prv, pub, in string) (string, error) {
	inByte, err := hex.DecodeString(in)
	if err != nil {
		return "", err
	}
	keyPrv, err := ecies.NewPrivateKeyFromHex(prv)
	if err != nil {
		return "", err
	}
	out, err := ecies.Decrypt(keyPrv, inByte)
	return string(out), err
}

func sign_ecies(prv, pub, in string) string {
	prvKey, err := ecies.NewPrivateKeyFromHex(prv)
	if err != nil {
		return ""
	}
	pubKey, err := ecies.NewPublicKeyFromHex(pub)
	if err != nil {
		return ""
	}
	log.Printf("ecies\nprv: %#+v\npub: %#+v\n", prvKey, pubKey)
	okey := &ecdsa.PrivateKey{
		D: prvKey.D,
		PublicKey: ecdsa.PublicKey{
			Curve: pubKey.Curve,
			X:     pubKey.X,
			Y:     pubKey.Y,
		},
	}
	log.Printf("ecdsa\nprv: %#+v\npub: %#+v\n", okey, okey.PublicKey)
	hash := sha3Of(in)
	signB, err := ecdsa.SignASN1(crand.Reader, okey, hash)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(signB)
}

func verify_ecies(pub, in, signature string) bool {
	pubKey, err := ecies.NewPublicKeyFromHex(pub)
	if err != nil {
		return false
	}
	okey := ecdsa.PublicKey{
		Curve: pubKey.Curve,
		X:     pubKey.X,
		Y:     pubKey.Y,
	}

	signB, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := sha3Of(in)
	return ecdsa.VerifyASN1(&okey, hash, signB)
}
