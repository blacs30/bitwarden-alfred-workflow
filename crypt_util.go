package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/hkdf"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func DecryptString(s string, mk CryptoKey) (string, error) {
	rv, err := DecryptValue(s, mk)
	return string(rv), err
}

func NewCryptoKey(key []byte, encryptionType int) (CryptoKey, error) {
	c := CryptoKey{EncryptionType: encryptionType}

	switch encryptionType {
	case AesCbc256_B64:
		c.EncKey = key
	case AesCbc256_HmacSha256_B64:
		c.EncKey = key[:32]
		c.MacKey = key[32:]
	default:
		return c, fmt.Errorf("invalid encryption type: %d", encryptionType)
	}

	if len(key) != (len(c.EncKey) + len(c.MacKey)) {
		return c, fmt.Errorf("invalid key size: %d", len(key))
	}

	return c, nil
}

func DecryptValue(s string, mk CryptoKey) ([]byte, error) {
	if s == "" {
		return []byte(""), nil
	}

	var rv []byte
	ck, err := NewCipherString(s)
	if err != nil {
		return rv, err
	}
	rv, err = ck.Decrypt(mk)
	return rv, err
}

func (cs *CipherString) DecryptKey(key CryptoKey, encryptionType int) (CryptoKey, error) {
	kb, err := cs.Decrypt(key)
	if err != nil {
		return CryptoKey{}, err
	}
	k, err := NewCryptoKey(kb, encryptionType)
	return k, err
}

func MakeIntermediateKeys(sourceKey CryptoKey) (CryptoKey, error) {
	tmpKeyEnc := make([]byte, 32)
	tmpKeyMac := make([]byte, 32)
	var r io.Reader
	r = hkdf.Expand(sha256.New, sourceKey.EncKey, []byte("enc"))
	_, err := r.Read(tmpKeyEnc)
	if err != nil {
		return CryptoKey{}, err
	}

	r = hkdf.Expand(sha256.New, sourceKey.EncKey, []byte("mac"))
	_, err = r.Read(tmpKeyMac)
	if err != nil {
		return CryptoKey{}, err
	}

	ck := CryptoKey{
		EncKey:         tmpKeyEnc,
		MacKey:         tmpKeyMac,
		EncryptionType: 2,
	}
	return ck, nil
}

func NewCipherString(encryptedString string) (*CipherString, error) {
	cs := CipherString{}
	cs.encryptedString = encryptedString
	if encryptedString == "" {
		return nil, errors.New("empty key")
	}
	headerPieces := strings.Split(cs.encryptedString, ".")
	var encPieces []string
	if len(headerPieces) == 2 {
		cs.encryptionType, _ = strconv.Atoi(headerPieces[0])
		encPieces = strings.Split(headerPieces[1], "|")
	} else {
		return nil, errors.New("invalid key header")
	}

	debugLog(fmt.Sprintf("cs.encryptionType %d", cs.encryptionType))
	switch cs.encryptionType {
	case AesCbc256_B64:
		if len(encPieces) != 2 {
			return nil, fmt.Errorf("invalid key body len %d", len(encPieces))
		}
		cs.initializationVector = encPieces[0]
		cs.cipherText = encPieces[1]
	case AesCbc256_HmacSha256_B64:
		if len(encPieces) != 3 {
			return nil, fmt.Errorf("invalid key body len %d", len(encPieces))
		}
		cs.initializationVector = encPieces[0]
		cs.cipherText = encPieces[1]
		cs.mac = encPieces[2]
	default:
		return nil, errors.New("unknown algorithm")
	}
	return &cs, nil
}

func (cs *CipherString) Decrypt(key CryptoKey) ([]byte, error) {
	iv, err := base64.StdEncoding.DecodeString(cs.initializationVector)
	if err != nil {
		return nil, err
	}

	ct, err := base64.StdEncoding.DecodeString(cs.cipherText)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key.EncKey)
	if err != nil {
		return nil, err
	}

	if cs.mac != "" {
		mac := hmac.New(sha256.New, key.MacKey)
		_, err = mac.Write(iv)
		if err != nil {
			return nil, err
		}
		_, err = mac.Write(ct)
		if err != nil {
			return nil, err
		}
		ms := mac.Sum(nil)
		if base64.StdEncoding.EncodeToString(ms) != cs.mac {
			return ct, fmt.Errorf("MAC doesn't match %s %s", cs.mac, base64.StdEncoding.EncodeToString(ms))
		}
	}

	decrypter := cipher.NewCBCDecrypter(block, iv)
	dst := make([]byte, len(ct))
	decrypter.CryptBlocks(dst, ct)
	dst = unpad(dst)
	return dst, nil
}

func unpad(src []byte) []byte {
	n := src[len(src)-1]
	return src[:len(src)-int(n)]
}

// TOTP related functions
func otpKey(key string) (string, error) {
	removedWhitespace := strings.ReplaceAll(key, " ", "")
	fetchSecret := getTotpSecretFromString(removedWhitespace)
	code, err := totp.GenerateCode(fetchSecret, time.Now())
	if err != nil {
		return "", fmt.Errorf("Error generating totp code, %s", err)
	}
	log.Print("totp code: ", code)
	return code, nil
}

func getTotpSecretFromString(key string) string {
	// need to get the secret
	//?secret=JBSWY3DPE&is
	re := regexp.MustCompile("secret=(.*?)&")
	matches := re.MatchString(key)
	if matches {
		res := re.FindAllStringSubmatch(key, 1)
		log.Print(res[0][1])
		return res[0][1]
	}
	re = regexp.MustCompile("secret=(.*)")
	matches = re.MatchString(key)
	if matches {
		res := re.FindAllStringSubmatch(key, 1)
		log.Print(res[0][1])
		return res[0][1]
	}
	// no matches, return original key
	return key
}
