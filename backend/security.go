package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"

	"golang.org/x/crypto/chacha20"
)

// Generate a new security key.
func getNewSecurityKey(seed int) (string, error) {
	// Generate key.
	rand.Seed(int64(seed))
	keyNum := rand.Uint32()
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key[0:4], uint32(keyNum))

	// Generate nonce.
	rand.Seed(int64(seed + 1))
	nonceNum := int16(rand.Uint32())
	nonce := make([]byte, 2)
	binary.BigEndian.PutUint16(nonce[0:2], uint16(nonceNum))

	// Generate password, then base64 string.
	cipher, err := chacha20.HChaCha20(key, nonce)
	if err != nil {
		return "", err
	} else {
		return base64.StdEncoding.EncodeToString(cipher), nil
	}
}

// Check needed configuration setup before running server
func securityCheck() {
	// Check security token existence before running.
}
