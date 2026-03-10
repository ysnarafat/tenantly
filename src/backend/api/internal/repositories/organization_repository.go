package repositories

import (
	"database/sql"
	"fmt"

	"github.com/ysnarafat/tenantly/internal/models"
)

// OrganizationRepository handles database operations for organizations
type OrganizationRepository struct {
	db *sql.DB
}

// NewOrganizationRepository creates a new instance of OrganizationRepository
func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create inserts a new organization into the database
func (r *OrganizationRepository) Create(org *models.Organization) error {
	query := `
		INSERT INTO organizations (name, slug, subscription_tier, max_users, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		org.Name,
		org.Slug,
		org.SubscriptionTier,
		org.MaxUsers,
		org.Active,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

// GetByID retrieves an organization by its ID
func (r *OrganizationRepository) GetByID(id int) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, subscription_tier, max_users, active, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	org := &models.Organization{}
	err := r.db.QueryRow(query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.SubscriptionTier,
		&org.MaxUsers,
		&org.Active,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// GetBySlug retrieves an organization by its slug
func (r *OrganizationRepository) GetBySlug(slug string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, subscription_tier, max_users, active, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`

	org := &models.Organization{}
	err := r.db.QueryRow(query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.SubscriptionTier,
		&org.MaxUsers,
		&org.Active,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization with slug '%s' not found", slug)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by slug: %w", err)
	}

	return org, nil
}

// GetAll retrieves all organizations, optionally filtering by active status
func (r *OrganizationRepository) GetAll(activeOnly bool) ([]*models.Organization, error) {
	query := `
		SELECT id, name, slug, subscription_tier, max_users, active, created_at, updated_at
		FROM organizations
	`

	args := []interface{}{}

	if activeOnly {
		query += " WHERE active = $1"
		args = append(args, true)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.SubscriptionTier,
			&org.MaxUsers,
			&org.Active,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, org)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organizations: %w", err)
	}

	return organizations, nil
}

// Update modifies an existing organization (soft update)
func (r *OrganizationRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Allowed fields for updates
	allowedFields := map[string]bool{
		"name":                 true,
		"slug":                 true,
		"subscription_tier":    true,
		"max_users":            true,
		"active":               true,
	}

	// Build the query dynamically
	query := "UPDATE organizations SET "
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("field '%s' cannot be updated", field)
		}
		if argIndex > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	args = append(args, id)
	query += fmt.Sprintf(" WHERE id = $%d", argIndex)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// Delete performs a soft delete of an organization (sets active = false)
func (r *OrganizationRepository) Delete(id int) error {
	query := "UPDATE organizations SET active = false WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}
