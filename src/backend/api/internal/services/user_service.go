package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo      interfaces.UserRepositoryInterface
	auditService  interfaces.AuditServiceInterface
	jwtSecret     string
	jwtExpiration time.Duration
}

func NewUserService(userRepo interfaces.UserRepositoryInterface, auditService interfaces.AuditServiceInterface, jwtSecret string, jwtExpiration time.Duration) *UserService {
	return &UserService{
		userRepo:      userRepo,
		auditService:  auditService,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Password validation constants
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

// ValidatePassword checks if password meets security requirements
func (s *UserService) ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must be no more than %d characters long", MaxPasswordLength)
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateUserInput validates user input data
func (s *UserService) ValidateUserInput(req *models.CreateUserRequest) error {
	// Validate username
	if len(strings.TrimSpace(req.Username)) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(req.Username) > 50 {
		return fmt.Errorf("username must be no more than 50 characters long")
	}

	// Check for valid username characters (alphanumeric and underscore only)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validUsername.MatchString(req.Username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	// Validate email format (additional validation beyond binding)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate role
	validRoles := map[string]bool{
		"Admin":           true,
		"PropertyManager": true,
		"Accountant":      true,
	}
	if !validRoles[req.Role] {
		return fmt.Errorf("invalid role: must be Admin, PropertyManager, or Accountant")
	}

	return nil
}

func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Validate input data
	if err := s.ValidateUserInput(req); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Validate password strength
	if err := s.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Normalize input
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if username already exists
	existingUser, _ := s.userRepo.GetByUsername(username)
	if existingUser != nil {
		return nil, fmt.Errorf("username '%s' already exists", username)
	}

	// Check if email already exists
	existingUserByEmail, _ := s.userRepo.GetByEmail(email)
	if existingUserByEmail != nil {
		return nil, fmt.Errorf("email '%s' already exists", email)
	}

	// Hash password with higher cost for better security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Active:       true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log user creation
	if s.auditService != nil {
		s.auditService.LogSystemAction(
			models.AuditActionCreate,
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		)
	}

	return user, nil
}

func (s *UserService) Login(req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error) {

	// Normalize username
	username := strings.TrimSpace(req.Username)

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("no user found with that username")
	}

	// Check if user is active
	if !user.Active {
		// Log failed login attempt for inactive user
		if s.auditService != nil {
			s.auditService.LogUserAction(
				user.ID,
				models.AuditActionLogin,
				models.TableUsers,
				&user.ID,
				nil,
				map[string]interface{}{
					"username":   username,
					"success":    false,
					"reason":     "user_inactive",
					"ip_address": clientIP,
					"user_agent": userAgent,
				},
			)
		}
		return nil, fmt.Errorf("account is deactivated")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Log failed login attempt
		if s.auditService != nil {
			s.auditService.LogUserAction(
				user.ID,
				models.AuditActionLogin,
				models.TableUsers,
				&user.ID,
				nil,
				map[string]interface{}{
					"username":   username,
					"success":    false,
					"reason":     "invalid_password",
					"ip_address": clientIP,
					"user_agent": userAgent,
				},
			)
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Log successful login
	if s.auditService != nil {
		s.auditService.LogUserAction(
			user.ID,
			models.AuditActionLogin,
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"username":   username,
				"success":    true,
				"ip_address": clientIP,
				"user_agent": userAgent,
			},
		)
	}

	return &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresAt:    time.Now().Add(s.jwtExpiration),
	}, nil
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) UpdateUser(id int, req *models.UpdateUserRequest) error {
	// Check if user exists
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	updates := make(map[string]interface{})
	oldValues := make(map[string]interface{})

	// Validate and prepare username update
	if req.Username != "" {
		username := strings.TrimSpace(req.Username)

		// Validate username format
		if len(username) < 3 || len(username) > 50 {
			return fmt.Errorf("username must be between 3 and 50 characters")
		}

		validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !validUsername.MatchString(username) {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}

		// Check if username is already taken by another user
		if username != existingUser.Username {
			existingUserByUsername, _ := s.userRepo.GetByUsername(username)
			if existingUserByUsername != nil && existingUserByUsername.ID != id {
				return fmt.Errorf("username '%s' already exists", username)
			}
			oldValues["username"] = existingUser.Username
			updates["username"] = username
		}
	}

	// Validate and prepare email update
	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))

		// Validate email format
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(email) {
			return fmt.Errorf("invalid email format")
		}

		// Check if email is already taken by another user
		if email != existingUser.Email {
			existingUserByEmail, _ := s.userRepo.GetByEmail(email)
			if existingUserByEmail != nil && existingUserByEmail.ID != id {
				return fmt.Errorf("email '%s' already exists", email)
			}
			oldValues["email"] = existingUser.Email
			updates["email"] = email
		}
	}

	// Validate and prepare role update
	if req.Role != "" {
		validRoles := map[string]bool{
			"Admin":           true,
			"PropertyManager": true,
			"Accountant":      true,
		}
		if !validRoles[req.Role] {
			return fmt.Errorf("invalid role: must be Admin, PropertyManager, or Accountant")
		}

		if req.Role != existingUser.Role {
			oldValues["role"] = existingUser.Role
			updates["role"] = req.Role
		}
	}

	// Prepare active status update
	if req.Active != nil {
		if *req.Active != existingUser.Active {
			oldValues["active"] = existingUser.Active
			updates["active"] = *req.Active
		}
	}

	// If no changes, return success
	if len(updates) == 0 {
		return nil
	}

	// Perform update
	if err := s.userRepo.Update(id, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log user update
	if s.auditService != nil {
		s.auditService.LogSystemAction(
			models.AuditActionUpdate,
			models.TableUsers,
			&id,
			oldValues,
			updates,
		)
	}

	return nil
}

func (s *UserService) DeleteUser(id int) error {
	// Check if user exists
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Perform soft delete
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log user deletion
	if s.auditService != nil {
		s.auditService.LogSystemAction(
			models.AuditActionDelete,
			models.TableUsers,
			&id,
			map[string]interface{}{
				"username": existingUser.Username,
				"email":    existingUser.Email,
				"role":     existingUser.Role,
			},
			map[string]interface{}{
				"active": false,
			},
		)
	}

	return nil
}

// generateTokens creates both access and refresh tokens
func (s *UserService) generateTokens(user *models.User) (string, string, error) {
	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"type":     "access",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(s.jwtExpiration).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (longer expiration)
	refreshClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"type":     "refresh",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

// RefreshToken generates a new access token from a valid refresh token
func (s *UserService) RefreshToken(refreshTokenString string) (*models.LoginResponse, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get user ID
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Log token refresh
	if s.auditService != nil {
		s.auditService.LogUserAction(
			user.ID,
			"TOKEN_REFRESH",
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"username": user.Username,
			},
		)
	}

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
		ExpiresAt:    time.Now().Add(s.jwtExpiration),
	}, nil
}

// Logout logs out a user and optionally invalidates tokens
func (s *UserService) Logout(userID int, clientIP, userAgent string) error {
	// Log logout
	if s.auditService != nil {
		s.auditService.LogUserAction(
			userID,
			models.AuditActionLogout,
			models.TableUsers,
			&userID,
			nil,
			map[string]interface{}{
				"ip_address": clientIP,
				"user_agent": userAgent,
			},
		)
	}

	// In a production system, you might want to maintain a blacklist of tokens
	// or store active sessions in Redis/database for proper token invalidation
	return nil
}

// ChangePassword allows users to change their password
func (s *UserService) ChangePassword(userID int, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	updates := map[string]interface{}{
		"password_hash": string(hashedPassword),
		"updated_at":    time.Now(),
	}

	if err := s.userRepo.Update(userID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log password change
	if s.auditService != nil {
		s.auditService.LogUserAction(
			userID,
			"PASSWORD_CHANGE",
			models.TableUsers,
			&userID,
			nil,
			map[string]interface{}{
				"username": user.Username,
			},
		)
	}

	return nil
}

// ResetPassword generates a secure password reset token
func (s *UserService) ResetPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		// Still log the attempt for security monitoring
		if s.auditService != nil {
			s.auditService.LogSystemAction(
				"PASSWORD_RESET_REQUEST",
				models.TableUsers,
				nil,
				nil,
				map[string]interface{}{
					"email":   email,
					"success": false,
					"reason":  "email_not_found",
				},
			)
		}
		return nil
	}

	if !user.Active {
		// Don't reveal if user is inactive
		return nil
	}

	// Generate secure reset token
	resetToken, err := s.generateSecureToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create reset token record
	tokenRecord := &models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiration
		Used:      false,
	}

	if err := s.userRepo.CreateResetToken(tokenRecord); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Log password reset request
	if s.auditService != nil {
		s.auditService.LogUserAction(
			user.ID,
			"PASSWORD_RESET_REQUEST",
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"email":    email,
				"success":  true,
				"token_id": tokenRecord.ID,
			},
		)
	}

	// In a production system, you would send an email here
	// For now, we'll just log that the token was generated
	// TODO: Integrate with email service to send reset link

	return nil
}

// ConfirmPasswordReset validates reset token and updates password
func (s *UserService) ConfirmPasswordReset(token, newPassword string) error {
	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Get and validate reset token
	resetToken, err := s.userRepo.GetResetToken(token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Get user
	user, err := s.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !user.Active {
		return fmt.Errorf("user account is deactivated")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	updates := map[string]interface{}{
		"password_hash": string(hashedPassword),
		"updated_at":    time.Now(),
	}

	if err := s.userRepo.Update(user.ID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.userRepo.MarkResetTokenUsed(resetToken.ID); err != nil {
		// Log error but don't fail the operation
		if s.auditService != nil {
			s.auditService.LogSystemAction(
				"PASSWORD_RESET_TOKEN_CLEANUP_FAILED",
				models.TableUsers,
				&user.ID,
				nil,
				map[string]interface{}{
					"token_id": resetToken.ID,
					"error":    err.Error(),
				},
			)
		}
	}

	// Log successful password reset
	if s.auditService != nil {
		s.auditService.LogUserAction(
			user.ID,
			"PASSWORD_RESET_COMPLETED",
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"username": user.Username,
				"token_id": resetToken.ID,
			},
		)
	}

	return nil
}

// generateSecureToken creates a cryptographically secure random token
func (s *UserService) generateSecureToken() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode as base64 URL-safe string
	return base64.URLEncoding.EncodeToString(bytes), nil
}
