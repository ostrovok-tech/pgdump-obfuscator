package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

var Salt []byte

func ScrambleEmail(s []byte) []byte {
	atIndex := bytes.IndexRune(s, '@')
	mailbox := Salt
	domain := []byte("@obfu.com")
	if atIndex != -1 {
		mailbox = s[:atIndex]
	}
	return append(ScrambleBytes(mailbox)[:13], domain...)
}

func ScramblePhone(s []byte) []byte {
	if len(s) == 1 {
		return s[:0]
	}

	hash := sha256.New()
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	// Exploit the fact that 32 bytes hash sum is greater than most phone numbers.
	s[0] = '+'
	for i := range s[1:] {
		if i+1 >= len(sumBytes) {
			break
		}
		s[i+1] = '0' + sumBytes[i]%10
	}
	return s
}

func GenScrambleBytes(maxLength uint) func([]byte) []byte {
	return func(s []byte) []byte {
		return ScrambleBytes(s)[:maxLength]
	}
}

func ScrambleBytes(s []byte) []byte {
	hash := sha256.New()
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	b64 := make([]byte, base64.URLEncoding.EncodedLen(len(sumBytes)))
	base64.URLEncoding.Encode(b64, sumBytes)
	return b64
}

func init() {
	Salt = make([]byte, 16)
	_, err := rand.Read(Salt)
	if err != nil {
		panic(err)
	}
}
