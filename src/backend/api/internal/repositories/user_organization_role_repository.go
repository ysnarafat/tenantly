package repositories

import (
	"database/sql"
	"fmt"

	"github.com/ysnarafat/tenantly/internal/models"
)

type UserOrganizationRoleRepository struct {
	db *sql.DB
}

func NewUserOrganizationRoleRepository(db *sql.DB) *UserOrganizationRoleRepository {
	return &UserOrganizationRoleRepository{db: db}
}

// GetByUserID returns all organization roles for a user, joining organization data
func (r *UserOrganizationRoleRepository) GetByUserID(userID int) ([]models.UserOrganizationRole, error) {
	query := `
		SELECT
			uor.id, uor.user_id, uor.organization_id, uor.role,
			uor.created_at, uor.updated_at,
			o.id, o.name, o.slug, o.subscription_tier, o.max_users, o.active, o.created_at, o.updated_at
		FROM user_organization_roles uor
		INNER JOIN organizations o ON o.id = uor.organization_id
		WHERE uor.user_id = $1 AND o.active = true
		ORDER BY o.name ASC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user organizations: %w", err)
	}
	defer rows.Close()

	var roles []models.UserOrganizationRole
	for rows.Next() {
		var uor models.UserOrganizationRole
		err := rows.Scan(
			&uor.ID,
			&uor.UserID,
			&uor.OrganizationID,
			&uor.Role,
			&uor.CreatedAt,
			&uor.UpdatedAt,
			&uor.Organization.ID,
			&uor.Organization.Name,
			&uor.Organization.Slug,
			&uor.Organization.SubscriptionTier,
			&uor.Organization.MaxUsers,
			&uor.Organization.Active,
			&uor.Organization.CreatedAt,
			&uor.Organization.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user organization role: %w", err)
		}
		roles = append(roles, uor)
	}

	return roles, rows.Err()
}

// GetByUserAndOrg returns the role for a specific user-organization pair
func (r *UserOrganizationRoleRepository) GetByUserAndOrg(userID, orgID int) (*models.UserOrganizationRole, error) {
	query := `
		SELECT
			uor.id, uor.user_id, uor.organization_id, uor.role,
			uor.created_at, uor.updated_at,
			o.id, o.name, o.slug, o.subscription_tier, o.max_users, o.active, o.created_at, o.updated_at
		FROM user_organization_roles uor
		INNER JOIN organizations o ON o.id = uor.organization_id
		WHERE uor.user_id = $1 AND uor.organization_id = $2 AND o.active = true`

	var uor models.UserOrganizationRole
	err := r.db.QueryRow(query, userID, orgID).Scan(
		&uor.ID,
		&uor.UserID,
		&uor.OrganizationID,
		&uor.Role,
		&uor.CreatedAt,
		&uor.UpdatedAt,
		&uor.Organization.ID,
		&uor.Organization.Name,
		&uor.Organization.Slug,
		&uor.Organization.SubscriptionTier,
		&uor.Organization.MaxUsers,
		&uor.Organization.Active,
		&uor.Organization.CreatedAt,
		&uor.Organization.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user does not have access to this organization")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user organization role: %w", err)
	}

	return &uor, nil
}

// Upsert inserts or updates a user-organization role
func (r *UserOrganizationRoleRepository) Upsert(userID, orgID int, role string) error {
	query := `
		INSERT INTO user_organization_roles (user_id, organization_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, organization_id) DO UPDATE SET role = $3, updated_at = NOW()`

	_, err := r.db.Exec(query, userID, orgID, role)
	if err != nil {
		return fmt.Errorf("failed to upsert user organization role: %w", err)
	}
	return nil
}

// Delete removes a user from an organization
func (r *UserOrganizationRoleRepository) Delete(userID, orgID int) error {
	query := `DELETE FROM user_organization_roles WHERE user_id = $1 AND organization_id = $2`
	_, err := r.db.Exec(query, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete user organization role: %w", err)
	}
	return nil
}
