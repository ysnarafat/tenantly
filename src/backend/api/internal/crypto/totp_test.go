package crypto

import (
	"strings"
	"testing"
)

func TestGenerateTOTPSecret(t *testing.T) {
	s, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if s == "" || strings.Contains(s, "=") {
		t.Fatalf("unexpected secret %q", s)
	}
}

func TestValidateTOTP_RoundTrip(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	now := int64(1_700_000_000)

	code, err := hotp(secret, now/totpPeriod)
	if err != nil {
		t.Fatalf("hotp failed: %v", err)
	}
	if len(code) != totpDigits {
		t.Fatalf("expected %d digits, got %q", totpDigits, code)
	}
	if !ValidateTOTP(secret, code, now) {
		t.Fatal("expected current code to validate")
	}
}

func TestValidateTOTP_Skew(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	now := int64(1_700_000_000)

	prev, _ := hotp(secret, now/totpPeriod-1)
	next, _ := hotp(secret, now/totpPeriod+1)
	if !ValidateTOTP(secret, prev, now) {
		t.Fatal("expected previous-step code to validate within skew")
	}
	if !ValidateTOTP(secret, next, now) {
		t.Fatal("expected next-step code to validate within skew")
	}

	far, _ := hotp(secret, now/totpPeriod+5)
	if ValidateTOTP(secret, far, now) {
		t.Fatal("expected far-future code to be rejected")
	}
}

func TestValidateTOTP_Rejects(t *testing.T) {
	secret, _ := GenerateTOTPSecret()
	if ValidateTOTP(secret, "000", 1_700_000_000) {
		t.Fatal("wrong-length code must be rejected")
	}
	if ValidateTOTP(secret, "999999", 1_700_000_000) {
		// Astronomically unlikely to be the real code.
		t.Fatal("arbitrary code must be rejected")
	}
}

func TestTOTPProvisioningURI(t *testing.T) {
	uri := TOTPProvisioningURI("ABC234", "Tenantly", "user@example.com")
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("unexpected URI %q", uri)
	}
	for _, want := range []string{"secret=ABC234", "issuer=Tenantly", "digits=6", "period=30"} {
		if !strings.Contains(uri, want) {
			t.Fatalf("URI %q missing %q", uri, want)
		}
	}
}
