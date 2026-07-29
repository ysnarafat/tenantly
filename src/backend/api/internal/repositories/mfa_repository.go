package repositories

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MFAEnrollment is a user's stored TOTP enrollment. Secret is the encrypted
// TOTP secret as persisted; callers decrypt it.
type MFAEnrollment struct {
	UserID  int
	Secret  string
	Enabled bool
}

// MFARepository persists per-user TOTP MFA enrollments.
type MFARepository struct {
	db *sqlx.DB
}

func NewMFARepository(db *sqlx.DB) *MFARepository {
	return &MFARepository{db: db}
}

// GetByUserID returns the enrollment for a user, or (nil, nil) if none exists.
func (r *MFARepository) GetByUserID(userID int) (*MFAEnrollment, error) {
	var e MFAEnrollment
	err := r.db.QueryRow(
		`SELECT user_id, secret, enabled FROM user_mfa WHERE user_id = $1`, userID,
	).Scan(&e.UserID, &e.Secret, &e.Enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get MFA enrollment: %w", err)
	}
	return &e, nil
}

// UpsertSecret stores (or replaces) a user's encrypted secret, resetting the
// enrollment to un-confirmed. Used when a user (re-)enrolls.
func (r *MFARepository) UpsertSecret(userID int, encryptedSecret string) error {
	_, err := r.db.Exec(`
		INSERT INTO user_mfa (user_id, secret, enabled)
		VALUES ($1, $2, false)
		ON CONFLICT (user_id)
		DO UPDATE SET secret = EXCLUDED.secret, enabled = false, updated_at = NOW() AT TIME ZONE 'UTC'`,
		userID, encryptedSecret)
	if err != nil {
		return fmt.Errorf("failed to upsert MFA secret: %w", err)
	}
	return nil
}

// SetEnabled marks a user's enrollment confirmed (or disables it).
func (r *MFARepository) SetEnabled(userID int, enabled bool) error {
	_, err := r.db.Exec(
		`UPDATE user_mfa SET enabled = $2, updated_at = NOW() AT TIME ZONE 'UTC' WHERE user_id = $1`,
		userID, enabled)
	if err != nil {
		return fmt.Errorf("failed to set MFA enabled: %w", err)
	}
	return nil
}
