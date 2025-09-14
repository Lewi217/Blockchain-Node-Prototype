package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/ripemd160"
)

// SHA256 computes SHA-256 hash of input data
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA256Hex computes SHA-256 hash and returns hex string
func SHA256Hex(data []byte) string {
	hash := SHA256(data)
	return hex.EncodeToString(hash)
}

// DoubleHashSHA256 computes double SHA-256 hash (Bitcoin style)
func DoubleHashSHA256(data []byte) []byte {
	first := SHA256(data)
	second := SHA256(first)
	return second
}

// DoubleHashSHA256Hex computes double SHA-256 hash and returns hex string
func DoubleHashSHA256Hex(data []byte) string {
	hash := DoubleHashSHA256(data)
	return hex.EncodeToString(hash)
}

// RIPEMD160 computes RIPEMD-160 hash
func RIPEMD160(data []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// Hash160 computes SHA-256 followed by RIPEMD-160 (used for addresses)
func Hash160(data []byte) []byte {
	sha256Hash := SHA256(data)
	return RIPEMD160(sha256Hash)
}

// Hash160Hex computes Hash160 and returns hex string
func Hash160Hex(data []byte) string {
	hash := Hash160(data)
	return hex.EncodeToString(hash)
}