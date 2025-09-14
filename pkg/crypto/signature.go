package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Signature represents a digital signature
type Signature struct {
	R *big.Int
	S *big.Int
}

// Sign creates a digital signature for the given data using private key
func Sign(data []byte, privateKey *ecdsa.PrivateKey) (*Signature, error) {
	// Hash the data first
	hash := SHA256(data)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %v", err)
	}
	
	return &Signature{R: r, S: s}, nil
}

// SignHex creates a digital signature and returns it as hex string
func SignHex(data []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	signature, err := Sign(data, privateKey)
	if err != nil {
		return "", err
	}
	
	return signature.ToHex(), nil
}

// Verify verifies a digital signature against data and public key
func Verify(data []byte, signature *Signature, publicKey *ecdsa.PublicKey) bool {
	// Hash the data first
	hash := SHA256(data)
	
	// Verify the signature
	return ecdsa.Verify(publicKey, hash, signature.R, signature.S)
}

// VerifyHex verifies a hex-encoded signature
func VerifyHex(data []byte, signatureHex string, publicKey *ecdsa.PublicKey) bool {
	signature, err := SignatureFromHex(signatureHex)
	if err != nil {
		return false
	}
	
	return Verify(data, signature, publicKey)
}

// ToHex converts signature to hex string
func (s *Signature) ToHex() string {
	// Convert R and S to 32-byte arrays
	rBytes := s.R.Bytes()
	sBytes := s.S.Bytes()
	
	// Pad to 32 bytes if necessary
	for len(rBytes) < 32 {
		rBytes = append([]byte{0}, rBytes...)
	}
	for len(sBytes) < 32 {
		sBytes = append([]byte{0}, sBytes...)
	}
	
	// Combine R and S
	sigBytes := append(rBytes, sBytes...)
	return hex.EncodeToString(sigBytes)
}

// SignatureFromHex creates a signature from hex string
func SignatureFromHex(hexSig string) (*Signature, error) {
	sigBytes, err := hex.DecodeString(hexSig)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %v", err)
	}
	
	if len(sigBytes) != 64 {
		return nil, fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(sigBytes))
	}
	
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	
	return &Signature{R: r, S: s}, nil
}

// String returns string representation of signature
func (s *Signature) String() string {
	return fmt.Sprintf("Signature{R: %s, S: %s}", s.R.String(), s.S.String())
}