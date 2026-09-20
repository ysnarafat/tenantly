package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

// StepUpPurpose identifies the short-lived token minted after a successful MFA
// challenge. It is distinct from the normal access token so a stolen access
// token alone cannot satisfy a step-up-gated action.
const StepUpPurpose = "mfa_stepup"

// stepUpTTL is how long an MFA verification remains valid for gated actions.
const stepUpTTL = 5 * time.Minute

// MFAService handles TOTP enrollment, verification, and step-up token minting.
type MFAService struct {
	repo      *repositories.MFARepository
	protector *appcrypto.NIDProtector
	jwtSecret string
	issuer    string
}

func NewMFAService(repo *repositories.MFARepository, protector *appcrypto.NIDProtector, jwtSecret string) *MFAService {
	return &MFAService{repo: repo, protector: protector, jwtSecret: jwtSecret, issuer: "Tenantly"}
}

// Status reports whether the user has a confirmed MFA enrollment.
func (s *MFAService) Status(userID int) (bool, error) {
	enrollment, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, err
	}
	return enrollment != nil && enrollment.Enabled, nil
}

// Enroll generates a fresh TOTP secret for the user, stores it encrypted (as
// un-confirmed), and returns the secret plus the otpauth provisioning URI for
// the client to render as a QR code. Enrollment is completed by Verify.
func (s *MFAService) Enroll(userID int, account string) (secret, uri string, err error) {
	secret, err = appcrypto.GenerateTOTPSecret()
	if err != nil {
		return "", "", err
	}
	encrypted, err := s.protector.Encrypt(secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt MFA secret: %w", err)
	}
	if err := s.repo.UpsertSecret(userID, encrypted); err != nil {
		return "", "", err
	}
	return secret, appcrypto.TOTPProvisioningURI(secret, s.issuer, account), nil
}

// Verify validates a TOTP code. On the first successful verification it confirms
// enrollment. On success it returns a short-lived step-up token and its expiry.
func (s *MFAService) Verify(userID int, code string) (stepUpToken string, expiresAt time.Time, err error) {
	enrollment, err := s.repo.GetByUserID(userID)
	if err != nil {
		return "", time.Time{}, err
	}
	if enrollment == nil {
		return "", time.Time{}, fmt.Errorf("MFA not enrolled")
	}

	secret, err := s.protector.Decrypt(enrollment.Secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !appcrypto.ValidateTOTP(secret, code, time.Now().Unix()) {
		return "", time.Time{}, fmt.Errorf("invalid code")
	}

	if !enrollment.Enabled {
		if err := s.repo.SetEnabled(userID, true); err != nil {
			return "", time.Time{}, err
		}
	}

	expiresAt = time.Now().Add(stepUpTTL)
	claims := jwt.MapClaims{
		"user_id": userID,
		"purpose": StepUpPurpose,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	stepUpToken, err = token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign step-up token: %w", err)
	}
	return stepUpToken, expiresAt, nil
}
