package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	AddressVersionMainNet = 0x00 
	AddressVersionTestNet = 0x6f 
)

func GenerateAddress(publicKey *ecdsa.PublicKey, version byte) string {

	pubKeyBytes := compressPublicKey(publicKey)
	
	hash160 := Hash160(pubKeyBytes)
	
	versionedHash := append([]byte{version}, hash160...)
	
	checksum := calculateChecksum(versionedHash)
	
	fullAddress := append(versionedHash, checksum...)
	
	return hex.EncodeToString(fullAddress)
}

func GenerateAddressFromKeyPair(kp *KeyPair, version byte) string {
	return GenerateAddress(kp.PublicKey, version)
}

func ValidateAddress(address string) bool {
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return false
	}
	
	if len(addressBytes) != 25 { 
		return false
	}
	
	versionedHash := addressBytes[:21]
	checksum := addressBytes[21:]
	
	expectedChecksum := calculateChecksum(versionedHash)
	
	for i := 0; i < 4; i++ {
		if checksum[i] != expectedChecksum[i] {
			return false
		}
	}
	
	return true
}

func ExtractPublicKeyHash(address string) ([]byte, error) {
	if !ValidateAddress(address) {
		return nil, fmt.Errorf("invalid address")
	}
	
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	
	return addressBytes[1:21], nil
}

func compressPublicKey(publicKey *ecdsa.PublicKey) []byte {
	x := publicKey.X.Bytes()
	
	for len(x) < 32 {
		x = append([]byte{0}, x...)
	}
	
	var prefix byte
	if publicKey.Y.Bit(0) == 0 {
		prefix = 0x02
	} else {
		prefix = 0x03 
	}
	
	return append([]byte{prefix}, x...)
}


func calculateChecksum(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	
	return second[:4]
}

func IsValidAddressForNetwork(address string, version byte) bool {
	if !ValidateAddress(address) {
		return false
	}
	
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return false
	}
	
	return addressBytes[0] == version
}

func AddressToString(addressBytes []byte) string {
	return hex.EncodeToString(addressBytes)
}

func StringToAddress(addressStr string) ([]byte, error) {
	return hex.DecodeString(addressStr)
}