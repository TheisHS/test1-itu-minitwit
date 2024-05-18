package main

import (
	"fmt"

	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"

	"encoding/hex"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Iterations = 150000
	satlChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLength = 32
)

func GeneratePasswordHash(password string) string {
	if password == "" { return "" }
	salt := genSalt()
	hash := hashString(salt, password)
	method := "pbkdf2:sha256"
	return fmt.Sprintf("%s:%v$%s$%s", method, Iterations, salt, hash)
}

func genSalt() string {
	saltLength := 8
	var bytes = make([]byte, saltLength)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = satlChars[v%byte(len(satlChars))]
	}
	return string(bytes)
}

func hashString(salt string, password string) string {
	hash := pbkdf2.Key([]byte(password), []byte(salt), Iterations, keyLength, sha256.New)
	return hex.EncodeToString(hash)
}