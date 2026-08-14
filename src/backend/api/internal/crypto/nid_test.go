package crypto

import (
	"crypto/sha256"
	"testing"
)

func newTestProtector(t *testing.T) *NIDProtector {
	t.Helper()
	key := sha256.Sum256([]byte("nid-crypto-test-key"))
	p, err := NewNIDProtector(key[:], []byte("nid-crypto-test-pepper"))
	if err != nil {
		t.Fatalf("failed to build protector: %v", err)
	}
	return p
}

func TestNewNIDProtector_KeyValidation(t *testing.T) {
	if _, err := NewNIDProtector(make([]byte, 16), []byte("p")); err == nil {
		t.Fatal("expected error for short key")
	}
	if _, err := NewNIDProtector(make([]byte, 32), nil); err == nil {
		t.Fatal("expected error for empty pepper")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	p := newTestProtector(t)
	plaintext := "1990123456789"

	encrypted, err := p.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if encrypted == plaintext {
		t.Fatal("ciphertext must not equal plaintext")
	}

	decrypted, err := p.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("round-trip mismatch: got %q want %q", decrypted, plaintext)
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	p := newTestProtector(t)
	a, _ := p.Encrypt("1234567890")
	b, _ := p.Encrypt("1234567890")
	if a == b {
		t.Fatal("expected distinct ciphertexts due to random nonce")
	}
}

func TestEncryptDecrypt_Empty(t *testing.T) {
	p := newTestProtector(t)
	enc, err := p.Encrypt("")
	if err != nil || enc != "" {
		t.Fatalf("empty encrypt should yield empty string, got %q err %v", enc, err)
	}
	dec, err := p.Decrypt("")
	if err != nil || dec != "" {
		t.Fatalf("empty decrypt should yield empty string, got %q err %v", dec, err)
	}
}

func TestDecrypt_WrongKeyFails(t *testing.T) {
	p := newTestProtector(t)
	encrypted, _ := p.Encrypt("1234567890")

	otherKey := sha256.Sum256([]byte("different-key"))
	other, _ := NewNIDProtector(otherKey[:], []byte("other-pepper"))
	if _, err := other.Decrypt(encrypted); err == nil {
		t.Fatal("expected decryption with wrong key to fail (GCM auth)")
	}
}

func TestHash_DeterministicAndKeyed(t *testing.T) {
	p := newTestProtector(t)
	h1 := p.Hash("1234567890")
	h2 := p.Hash("1234567890")
	if h1 != h2 {
		t.Fatal("hash must be deterministic for lookup")
	}
	if h1 == "1234567890" || len(h1) != 64 {
		t.Fatalf("unexpected hash output %q", h1)
	}

	otherKey := sha256.Sum256([]byte("k"))
	other, _ := NewNIDProtector(otherKey[:], []byte("other-pepper"))
	if other.Hash("1234567890") == h1 {
		t.Fatal("different pepper must produce a different hash")
	}
	if p.Hash("") != "" {
		t.Fatal("empty plaintext should hash to empty string")
	}
}

func TestLastFourAndMask(t *testing.T) {
	if LastFour("1234567890") != "7890" {
		t.Fatal("LastFour should return last 4 characters")
	}
	if LastFour("12") != "12" {
		t.Fatal("LastFour of short value should return the whole value")
	}
	if Mask("1234567890") != "••••••7890" {
		t.Fatalf("unexpected mask %q", Mask("1234567890"))
	}
	if Mask("") != "" {
		t.Fatal("mask of empty should be empty")
	}
}
