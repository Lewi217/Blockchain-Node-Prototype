package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// KeyPair represents a public/private key pair
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// GenerateKeyPair generates a new ECDSA key pair using secp256k1 curve
func GenerateKeyPair() (*KeyPair, error) {
	// Use secp256r1 (P-256) curve (secp256k1 not in standard lib)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// GetPrivateKeyHex returns the private key as hex string
func (kp *KeyPair) GetPrivateKeyHex() string {
	return hex.EncodeToString(kp.PrivateKey.D.Bytes())
}

// GetPublicKeyHex returns the public key as hex string (compressed format)
func (kp *KeyPair) GetPublicKeyHex() string {
	// Compressed public key format: 0x02/0x03 + x coordinate
	x := kp.PublicKey.X.Bytes()
	
	// Pad x to 32 bytes if necessary
	for len(x) < 32 {
		x = append([]byte{0}, x...)
	}
	
	var prefix byte
	if kp.PublicKey.Y.Bit(0) == 0 {
		prefix = 0x02 // Even Y coordinate
	} else {
		prefix = 0x03 // Odd Y coordinate
	}
	
	compressed := append([]byte{prefix}, x...)
	return hex.EncodeToString(compressed)
}

// GetPublicKeyUncompressed returns the uncompressed public key as hex string
func (kp *KeyPair) GetPublicKeyUncompressed() string {
	// Uncompressed format: 0x04 + x coordinate + y coordinate
	x := kp.PublicKey.X.Bytes()
	y := kp.PublicKey.Y.Bytes()
	
	// Pad to 32 bytes if necessary
	for len(x) < 32 {
		x = append([]byte{0}, x...)
	}
}