package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// UserInvitationRepository handles database operations for user invitations
type UserInvitationRepository struct {
	db *sql.DB
}

// NewUserInvitationRepository creates a new instance of UserInvitationRepository
func NewUserInvitationRepository(db *sql.DB) *UserInvitationRepository {
	return &UserInvitationRepository{db: db}
}

// Create inserts a new user invitation into the database
func (r *UserInvitationRepository) Create(invitation *models.UserInvitation) error {
	query := `
		INSERT INTO user_invitations (organization_id, email, role, invitation_token, expires_at, invited_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		invitation.OrganizationID,
		invitation.Email,
		invitation.Role,
		invitation.InvitationToken,
		invitation.ExpiresAt,
		invitation.InvitedByUserID,
	).Scan(&invitation.ID, &invitation.CreatedAt, &invitation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user invitation: %w", err)
	}

	return nil
}

// GetByID retrieves a user invitation by its ID
func (r *UserInvitationRepository) GetByID(id int) (*models.UserInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invitation_token, expires_at, accepted_at, accepted_by_user_id, invited_by_user_id, created_at, updated_at
		FROM user_invitations
		WHERE id = $1
	`

	invitation := &models.UserInvitation{}
	err := r.db.QueryRow(query, id).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.InvitationToken,
		&invitation.ExpiresAt,
		&invitation.AcceptedAt,
		&invitation.AcceptedByUserID,
		&invitation.InvitedByUserID,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user invitation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user invitation: %w", err)
	}

	return invitation, nil
}

// GetByToken retrieves a user invitation by its token
func (r *UserInvitationRepository) GetByToken(token string) (*models.UserInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invitation_token, expires_at, accepted_at, accepted_by_user_id, invited_by_user_id, created_at, updated_at
		FROM user_invitations
		WHERE invitation_token = $1
	`

	invitation := &models.UserInvitation{}
	err := r.db.QueryRow(query, token).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.InvitationToken,
		&invitation.ExpiresAt,
		&invitation.AcceptedAt,
		&invitation.AcceptedByUserID,
		&invitation.InvitedByUserID,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation with token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}

	return invitation, nil
}

// GetByEmailAndOrg retrieves a user invitation by email and organization
func (r *UserInvitationRepository) GetByEmailAndOrg(email string, orgID int) (*models.UserInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invitation_token, expires_at, accepted_at, accepted_by_user_id, invited_by_user_id, created_at, updated_at
		FROM user_invitations
		WHERE email = $1 AND organization_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	invitation := &models.UserInvitation{}
	err := r.db.QueryRow(query, email, orgID).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.InvitationToken,
		&invitation.ExpiresAt,
		&invitation.AcceptedAt,
		&invitation.AcceptedByUserID,
		&invitation.InvitedByUserID,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation by email and org: %w", err)
	}

	return invitation, nil
}

// GetPendingByOrganization retrieves all pending invitations for an organization
func (r *UserInvitationRepository) GetPendingByOrganization(orgID int) ([]*models.UserInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invitation_token, expires_at, accepted_at, accepted_by_user_id, invited_by_user_id, created_at, updated_at
		FROM user_invitations
		WHERE organization_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*models.UserInvitation
	for rows.Next() {
		invitation := &models.UserInvitation{}
		err := rows.Scan(
			&invitation.ID,
			&invitation.OrganizationID,
			&invitation.Email,
			&invitation.Role,
			&invitation.InvitationToken,
			&invitation.ExpiresAt,
			&invitation.AcceptedAt,
			&invitation.AcceptedByUserID,
			&invitation.InvitedByUserID,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invitations: %w", err)
	}

	return invitations, nil
}

// AcceptInvitation marks an invitation as accepted
func (r *UserInvitationRepository) AcceptInvitation(invitationID int, userID int) error {
	query := `
		UPDATE user_invitations
		SET accepted_at = $1, accepted_by_user_id = $2
		WHERE id = $3 AND accepted_at IS NULL AND expires_at > NOW()
	`

	result, err := r.db.Exec(query, time.Now(), userID, invitationID)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already accepted or expired")
	}

	return nil
}

// Delete removes an invitation (revoke)
func (r *UserInvitationRepository) Delete(id int) error {
	query := "DELETE FROM user_invitations WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found")
	}

	return nil
}

// CleanupExpiredInvitations removes expired invitations that were never accepted
func (r *UserInvitationRepository) CleanupExpiredInvitations() error {
	query := `
		DELETE FROM user_invitations
		WHERE accepted_at IS NULL AND expires_at < NOW()
	`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired invitations: %w", err)
	}

	return nil
}

// GetByOrganization retrieves all invitations for an organization (pending and accepted)
func (r *UserInvitationRepository) GetByOrganization(orgID int) ([]*models.UserInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invitation_token, expires_at, accepted_at, accepted_by_user_id, invited_by_user_id, created_at, updated_at
		FROM user_invitations
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*models.UserInvitation
	for rows.Next() {
		invitation := &models.UserInvitation{}
		err := rows.Scan(
			&invitation.ID,
			&invitation.OrganizationID,
			&invitation.Email,
			&invitation.Role,
			&invitation.InvitationToken,
			&invitation.ExpiresAt,
			&invitation.AcceptedAt,
			&invitation.AcceptedByUserID,
			&invitation.InvitedByUserID,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invitations: %w", err)
	}

	return invitations, nil
}
