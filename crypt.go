// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"log"
	"strings"
)

func Encrypt(message []byte) (string, bool) {
	var nonce [24]byte
	_, err := io.ReadAtLeast(rand.Reader, nonce[:], 24)
	if err != nil {
		log.Println(err)
	}
	var password [32]byte
	_, err = io.ReadAtLeast(rand.Reader, password[:], 32)
	if err != nil {
		log.Println(err)
	}
	encrypted := secretbox.Seal(nil, message, &nonce, &password)
	base64Enc := base64.StdEncoding.EncodeToString(password[:])
	err = wf.Keychain.Set("encryptPassword", base64Enc)
	if err != nil {
		log.Println(err)
	}

	if wf.Debug() {
		log.Printf("%T \n", encrypted)
	}
	enHex := fmt.Sprintf("%x:%x", nonce[:], encrypted)
	if wf.Debug() {
		fmt.Println("ENCRYPTED:", enHex)
	}
	err = wf.Cache.Store(CACHE_NAME, []byte(enHex))
	if err != nil {
		log.Println(err)
	}
	return enHex, true
}

func Decrypt() ([]byte, error) {
	log.Println("Decrypting data now.")
	encryptedHex, err := wf.Cache.Load(CACHE_NAME)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var nonce2 [24]byte
	parts := strings.SplitN(string(encryptedHex), ":", 2)
	if len(parts) < 2 {
		log.Println("expected nonce")
		return nil, errors.New("expected nonce")
	}
	//get nonce
	bs, err := hex.DecodeString(parts[0])
	if err != nil || len(bs) != 24 {
		log.Println("invalid nonce")
		return nil, errors.New("Invalid nonce")
	}
	copy(nonce2[:], bs)
	// get message
	bs, err = hex.DecodeString(parts[1])
	if err != nil {
		log.Println("invalid message")
		return nil, errors.New("Invalid message")
	}
	passwordBase64, err := wf.Keychain.Get("encryptPassword")
	if err != nil {
		log.Println(err)
	}
	decoded64, err := base64.StdEncoding.DecodeString(passwordBase64)
	if err != nil {
		log.Println(err)
	}
	var password [32]byte
	copy(password[:], decoded64)

	// you need the password to open the sealed secret box
	msg, ok := secretbox.Open(nil, bs, &nonce2, &password)
	if !ok {
		log.Print("invalid message")
	}
	return msg, nil
}

// These notes helped a lot https://github.com/attie/bitwarden-decrypt
// as well as this repo https://github.com/mvdan/bitw
// and https://github.com/philhug/bitwarden-client-go
const (
	AesCbc256_B64 = 0
	//AesCbc128_HmacSha256_B64          = 1
	AesCbc256_HmacSha256_B64 = 2
	//Rsa2048_OaepSha256_B64            = 3
	//Rsa2048_OaepSha1_B64              = 4
	//Rsa2048_OaepSha256_HmacSha256_B64 = 5
	//Rsa2048_OaepSha1_HmacSha256_B64   = 6
)

type CipherString struct {
	encryptedString      string
	encryptionType       int
	decryptedValue       string
	cipherText           string
	initializationVector string
	mac                  string
}

type CryptoKey struct {
	EncKey         []byte
	MacKey         []byte
	EncryptionType int
}

// TODO: split up into functions
func MakeDecryptKeyFromSession(protectedKey string, sessionKey string) (CryptoKey, error) {
	pt, err := base64.StdEncoding.DecodeString(protectedKey)
	if err != nil {
		log.Print("Error decoding protectedKey, ", err)
	}
	// following every step from here:
	// https://github.com/attie/bitwarden-decrypt#protected-session-data
	encryptionType := pt[:1]
	encryptionTypeInt := int(encryptionType[0])
	iv := pt[1:17]
	pkmac := pt[17:49]
	ct := pt[49:]

	// and now here:
	// https://github.com/attie/bitwarden-decrypt#derive-source-key-from-protected-session-data
	ses, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		log.Print("Error decoding sessionKey, ", err)
	}
	sesec := ses[:32]
	sesmac := ses[32:64]

	// the key which will be returned later, or empty in case of error
	ck := CryptoKey{}

	mac := hmac.New(sha256.New, sesmac)
	_, err = mac.Write(iv)
	if err != nil {
		return ck, err
	}
	_, err = mac.Write(ct)
	if err != nil {
		return ck, err
	}
	ms := mac.Sum(nil)
	if base64.StdEncoding.EncodeToString(ms) != base64.StdEncoding.EncodeToString(pkmac) {
		log.Printf("MAC doesn't match %s %s", base64.StdEncoding.EncodeToString(pkmac), base64.StdEncoding.EncodeToString(ms))
	}

	// and this one now:
	// https://github.com/attie/bitwarden-decrypt#decrypt
	cs := CipherString{
		encryptedString:      "",
		encryptionType:       encryptionTypeInt,
		decryptedValue:       "",
		cipherText:           base64.StdEncoding.EncodeToString(ct),
		initializationVector: base64.StdEncoding.EncodeToString(iv),
		mac:                  base64.StdEncoding.EncodeToString(pkmac),
	}

	ck = CryptoKey{
		EncKey:         sesec,
		MacKey:         sesmac,
		EncryptionType: 2,
	}
	sourceKey, err := cs.DecryptKey(ck, 2)
	if err != nil {
		log.Print("Error decrypting key, ", err)
	}

	// making now the intermediate keys:
	// https://github.com/attie/bitwarden-decrypt#derive-intermediate-keys-from-source-key
	interKeys, err := MakeIntermediateKeys(sourceKey)
	if err != nil {
		log.Print("Error making intermediate keys, ", err)
	}

	// finally decrypting the real users encryption key:
	// https://github.com/attie/bitwarden-decrypt#decrypt-the-users-final-keys
	ekCs, err := NewCipherString(bwData.EncKey)
	if err != nil {
		log.Print("Error making cipherstring from encKey, ", err)
	}
	userDecryptKey, err := ekCs.DecryptKey(interKeys, ekCs.encryptionType)
	if err != nil {
		log.Print("Error decrypting key, ", err)
	}
	tmpKeyEnc := userDecryptKey[:32]
	tmpKeyMac := userDecryptKey[32:64]
	userKey := CryptoKey{
		EncKey:         tmpKeyEnc,
		MacKey:         tmpKeyMac,
		EncryptionType: 2,
	}
	return userKey, err
}
