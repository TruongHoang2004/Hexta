// Package secret seals provider credentials at rest with AES-256-GCM.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrSealed = errors.New("credential unavailable")

// Sealer binds each ciphertext to caller-supplied associated data (for example a connection ID),
// so a ciphertext copied to another row fails to open.
type Sealer struct {
	Version int
	aead    cipher.AEAD
}

// NewSealer accepts a 32-byte key encoded as 64 hex characters.
func NewSealer(hexKey string, version int) (*Sealer, error) {
	key, err := hex.DecodeString(strings.TrimSpace(hexKey))
	if err != nil || len(key) != 32 || version < 1 {
		return nil, errors.New("invalid credential key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("invalid credential key")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("invalid credential key")
	}
	return &Sealer{Version: version, aead: aead}, nil
}

func (s *Sealer) Seal(plain, associated []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, ErrSealed
	}
	return s.aead.Seal(nonce, nonce, plain, associated), nil
}

func (s *Sealer) Open(sealed []byte, version int, associated []byte) ([]byte, error) {
	n := s.aead.NonceSize()
	if version != s.Version || len(sealed) <= n {
		return nil, ErrSealed
	}
	plain, err := s.aead.Open(nil, sealed[:n], sealed[n:], associated)
	if err != nil {
		return nil, ErrSealed
	}
	return plain, nil
}
