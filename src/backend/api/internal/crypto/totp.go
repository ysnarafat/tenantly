package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
)

const (
	totpPeriod = 30 // seconds per step (RFC 6238 default)
	totpDigits = 6
	totpSkew   = 1 // accept the adjacent step on each side for clock drift
)

// GenerateTOTPSecret returns a new random base32-encoded secret (no padding),
// suitable for an authenticator app.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20) // 160-bit secret, per RFC 4226 recommendation
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "="), nil
}

// TOTPProvisioningURI builds an otpauth:// URI that authenticator apps consume
// (renderable as a QR code by the client).
func TOTPProvisioningURI(secret, issuer, account string) string {
	label := url.PathEscape(issuer + ":" + account)
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", fmt.Sprintf("%d", totpDigits))
	q.Set("period", fmt.Sprintf("%d", totpPeriod))
	return fmt.Sprintf("otpauth://totp/%s?%s", label, q.Encode())
}

// ValidateTOTP reports whether code is valid for secret at the given Unix time,
// tolerating one step of clock skew on either side.
func ValidateTOTP(secret, code string, unixTime int64) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	counter := unixTime / totpPeriod
	for offset := int64(-totpSkew); offset <= totpSkew; offset++ {
		expected, err := hotp(secret, counter+offset)
		if err != nil {
			return false
		}
		// Constant-time comparison to avoid leaking timing information.
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// hotp computes an RFC 4226 HMAC-SHA1 one-time password for a counter.
func hotp(secret string, counter int64) (string, error) {
	key, err := base32.StdEncoding.DecodeString(padBase32(strings.ToUpper(secret)))
	if err != nil {
		return "", fmt.Errorf("invalid TOTP secret: %w", err)
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	value := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)

	mod := uint32(1)
	for i := 0; i < totpDigits; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", totpDigits, value%mod), nil
}

// padBase32 restores the '=' padding that GenerateTOTPSecret strips, so the
// standard decoder accepts the stored secret.
func padBase32(s string) string {
	if m := len(s) % 8; m != 0 {
		s += strings.Repeat("=", 8-m)
	}
	return s
}
