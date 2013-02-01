// Scramble functions.
// Input `s []byte` is required to be not nil.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
)

var Salt []byte

func scrambleOneEmail(s []byte) []byte {
	atIndex := bytes.IndexRune(s, '@')
	mailbox := Salt
	domain := []byte("@example.com")
	if atIndex != -1 {
		mailbox = s[:atIndex]
	}
	s = make([]byte, len(mailbox)+len(domain))
	copy(s, mailbox)
	copy(s[len(mailbox):], domain)
	// ScrambleBytes is in-place
	ScrambleBytes(s[0:len(mailbox)])
	return s
}

// Supports array of emails in format {email1,email2}
func ScrambleEmail(s []byte) []byte {
	if len(s) < 2 {
		// panic("ScrambleEmail: input is too small: '" + string(s) + "'")
		return s
	}
	if s[0] != '{' && s[len(s)-1] != '}' {
		return scrambleOneEmail(s)
	}
	parts := bytes.Split(s[1:len(s)-1], []byte{','})
	partsNew := make([][]byte, len(parts))
	outLength := 2 + len(parts) - 1
	for i, bs := range parts {
		partsNew[i] = scrambleOneEmail(bs)
		outLength += len(partsNew[i])
	}
	s = make([]byte, outLength)
	s[0] = '{'
	s[len(s)-1] = '}'
	copy(s[1:len(s)-1], bytes.Join(partsNew, []byte{','}))
	return s
}

func ScrambleDigits(s []byte) []byte {
	if len(s) == 1 {
		return s[:0]
	}

	hash := sha256.New()
	const sumLength = 32 // SHA256/8
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	for i, b := range s {
		if b >= '0' && b <= '9' {
			s[i] = '0' + (sumBytes[i%sumLength]+b)%10
		}
	}
	return s
}

func GenScrambleBytes(maxLength uint) func([]byte) []byte {
	return func(s []byte) []byte {
		// TODO: pad or extend s to maxLength
		return ScrambleBytes(s)[:maxLength]
	}
}

var bytesOutputAlphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+-_")
var bytesAlphabetLength = byte(len(bytesOutputAlphabet))

// Modifies `s` in-place.
func ScrambleBytes(s []byte) []byte {
	hash := sha256.New()
	// Hard-coding this constant wins less than 3% in BenchmarkScrambleBytes
	const sumLength = 32 // SHA256/8
	hash.Write(Salt)
	hash.Write(s)
	sumBytes := hash.Sum(nil)

	for i, b := range s {
		s[i] = bytesOutputAlphabet[(sumBytes[i%sumLength]+b)%bytesAlphabetLength]
	}
	return s
}

func init() {
	Salt = make([]byte, 16)
	_, err := rand.Read(Salt)
	if err != nil {
		panic(err)
	}
}
