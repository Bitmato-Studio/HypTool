package main

import (
	"crypto/rand"
	"math/big"
)

const ALPHABET = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const UUID_LEN = 10

func uuid() string {
	result := make([]byte, UUID_LEN)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(ALPHABET))))
		result[i] = ALPHABET[num.Int64()]
	}
	return string(result)
}
