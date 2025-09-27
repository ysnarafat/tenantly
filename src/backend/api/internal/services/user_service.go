package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo      *repositories.UserRepository
	auditService  *database.AuditService
	jwtSecret     string
	jwtExpiration time.Duration
}

func NewUserService(userRepo *repositories.UserRepository, auditService *database.AuditService, jwtSecret string, jwtExpiration time.Duration) *UserService {
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

func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Validate password strength
	if err := s.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if username already exists
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingUserByEmail, _ := s.userRepo.GetByEmail(req.Email)
	if existingUserByEmail != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password with higher cost for better security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
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
	updates := make(map[string]interface{})

	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	return s.userRepo.Update(id, updates)
}

func (s *UserService) DeleteUser(id int) error {
	return s.userRepo.Delete(id)
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

// ResetPassword generates a password reset token (simplified version)
func (s *UserService) ResetPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// In a real implementation, you would:
	// 1. Generate a secure reset token
	// 2. Store it in database with expiration
	// 3. Send email with reset link
	// For now, we'll just log the attempt

	if s.auditService != nil {
		s.auditService.LogUserAction(
			user.ID,
			"PASSWORD_RESET_REQUEST",
			models.TableUsers,
			&user.ID,
			nil,
			map[string]interface{}{
				"email": email,
			},
		)
	}

	return nil
}
