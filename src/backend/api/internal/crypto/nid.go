// Package crypto provides application-level protection for sensitive PII
// (National ID numbers) at rest: AES-256-GCM encryption for the full value,
// a deterministic keyed hash for uniqueness lookups without decryption, and
// a last-four projection for display/search.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// NIDProtector encrypts, hashes, and masks NID values. It is safe for
// concurrent use: the AEAD and HMAC key are immutable after construction.
type NIDProtector struct {
	aead   cipher.AEAD
	pepper []byte
}

// NewNIDProtector builds a protector from a 32-byte AES-256 key and a non-empty
// HMAC pepper. The pepper keys the deterministic lookup hash and must be kept
// secret; rotating it invalidates existing nid_hash values.
func NewNIDProtector(key, pepper []byte) (*NIDProtector, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("NID encryption key must be 32 bytes (got %d)", len(key))
	}
	if len(pepper) == 0 {
		return nil, fmt.Errorf("NID hash pepper must not be empty")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM AEAD: %w", err)
	}
	return &NIDProtector{aead: aead, pepper: pepper}, nil
}

// Encrypt returns a base64 string of nonce||ciphertext for the given plaintext.
// An empty plaintext encrypts to an empty string (no ciphertext stored).
func (p *NIDProtector) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	sealed := p.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. An empty input yields an empty string.
func (p *NIDProtector) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	nonceSize := p.aead.NonceSize()
	if len(sealed) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plaintext, err := p.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt NID: %w", err)
	}
	return string(plaintext), nil
}

// Hash returns a deterministic hex HMAC-SHA256 of the plaintext, keyed by the
// pepper. Deterministic (unlike bcrypt) so it can back an indexed uniqueness
// lookup; keyed so the stored digest is not attackable with a plain rainbow
// table. Empty plaintext yields an empty hash.
func (p *NIDProtector) Hash(plaintext string) string {
	if plaintext == "" {
		return ""
	}
	mac := hmac.New(sha256.New, p.pepper)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil))
}

// LastFour returns the last four characters of the plaintext (or the whole
// value if shorter), for non-sensitive display and search.
func LastFour(plaintext string) string {
	if len(plaintext) <= 4 {
		return plaintext
	}
	return plaintext[len(plaintext)-4:]
}

// Mask renders a value as bullets plus its last four characters, matching the
// frontend's presentation (e.g. "••••••6789").
func Mask(plaintext string) string {
	if plaintext == "" {
		return ""
	}
	if len(plaintext) <= 4 {
		return plaintext
	}
	return strings.Repeat("•", len(plaintext)-4) + plaintext[len(plaintext)-4:]
}
