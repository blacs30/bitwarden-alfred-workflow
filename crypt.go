// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"crypto/rand"
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
