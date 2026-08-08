package repositories

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, active, first_name, last_name, organization_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.Role, user.Active, user.FirstName, user.LastName, user.OrganizationID, user.Status).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, active, first_name, last_name, organization_id, status, created_at, updated_at
		FROM users WHERE id = $1 AND active = true`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.FirstName, &user.LastName, &user.OrganizationID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, active, first_name, last_name, organization_id, status, created_at, updated_at
		FROM users WHERE username = $1`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.FirstName, &user.LastName, &user.OrganizationID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, role, active, first_name, last_name, organization_id, status, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.FirstName, &user.LastName, &user.OrganizationID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetAll(activeOnly bool) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, active, first_name, last_name, organization_id, status, created_at, updated_at
		FROM users`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.Active, &user.FirstName, &user.LastName, &user.OrganizationID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		query += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) Delete(id int) error {
	query := "UPDATE users SET active = false, updated_at = NOW() WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) GetByOrganizationID(orgID int, activeOnly bool) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, active, first_name, last_name, organization_id, status, created_at, updated_at
		FROM users WHERE organization_id = $1`
	if activeOnly {
		query += ` AND active = true`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.Active, &user.FirstName, &user.LastName, &user.OrganizationID, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Password reset token methods
func (r *UserRepository) CreateResetToken(token *models.ResetPasswordToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, used)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.db.QueryRow(query, token.UserID, token.Token, token.ExpiresAt, token.Used).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *UserRepository) GetResetToken(token string) (*models.ResetPasswordToken, error) {
	resetToken := &models.ResetPasswordToken{}
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens 
		WHERE token = $1 AND used = false AND expires_at > NOW()`

	err := r.db.QueryRow(query, token).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token,
		&resetToken.ExpiresAt, &resetToken.Used, &resetToken.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return resetToken, nil
}

func (r *UserRepository) MarkResetTokenUsed(tokenID int) error {
	query := "UPDATE password_reset_tokens SET used = true WHERE id = $1"
	_, err := r.db.Exec(query, tokenID)
	return err
}

func (r *UserRepository) CleanupExpiredTokens() error {
	query := "DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used = true"
	_, err := r.db.Exec(query)
	return err
}
