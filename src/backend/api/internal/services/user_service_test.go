package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/models"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(id int, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateResetToken(token *models.ResetPasswordToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) GetResetToken(token string) (*models.ResetPasswordToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ResetPasswordToken), args.Error(1)
}

func (m *MockUserRepository) MarkResetTokenUsed(tokenID int) error {
	args := m.Called(tokenID)
	return args.Error(0)
}

func (m *MockUserRepository) CleanupExpiredTokens() error {
	args := m.Called()
	return args.Error(0)
}

// MockAuditService is a mock implementation of AuditServiceInterface
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	args := m.Called(userID, action, tableName, recordID, oldValues, newValues)
	return args.Error(0)
}

func (m *MockAuditService) LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	args := m.Called(action, tableName, recordID, oldValues, newValues)
	return args.Error(0)
}

func TestValidatePassword(t *testing.T) {
	userService := &UserService{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "StrongPass123!",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Short1!",
			wantErr:  true,
		},
		{
			name:     "No uppercase",
			password: "lowercase123!",
			wantErr:  true,
		},
		{
			name:     "No lowercase",
			password: "UPPERCASE123!",
			wantErr:  true,
		},
		{
			name:     "No digit",
			password: "NoDigitPass!",
			wantErr:  true,
		},
		{
			name:     "No special character",
			password: "NoSpecialChar123",
			wantErr:  true,
		},
		{
			name:     "Too long",
			password: "ThisPasswordIsWayTooLongAndExceedsTheMaximumLengthAllowedForPasswordsInTheSystemWhichShouldCauseValidationToFailBecauseItIsOver128Characters123!",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userService.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUserInput(t *testing.T) {
	userService := &UserService{}

	tests := []struct {
		name    string
		req     *models.CreateUserRequest
		wantErr bool
	}{
		{
			name: "Valid input",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "ValidPass123!",
				Role:     "Admin",
			},
			wantErr: false,
		},
		{
			name: "Username too short",
			req: &models.CreateUserRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "ValidPass123!",
				Role:     "Admin",
			},
			wantErr: true,
		},
		{
			name: "Invalid username characters",
			req: &models.CreateUserRequest{
				Username: "test-user",
				Email:    "test@example.com",
				Password: "ValidPass123!",
				Role:     "Admin",
			},
			wantErr: true,
		},
		{
			name: "Invalid email format",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "ValidPass123!",
				Role:     "Admin",
			},
			wantErr: true,
		},
		{
			name: "Invalid role",
			req: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "ValidPass123!",
				Role:     "InvalidRole",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userService.ValidateUserInput(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	userService := &UserService{
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "Admin",
	}

	accessToken, refreshToken, err := userService.generateTokens(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	req := &models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Role:     "Admin",
	}

	// Mock repository calls
	mockRepo.On("GetByUsername", "testuser").Return(nil, sql.ErrNoRows)
	mockRepo.On("GetByEmail", "test@example.com").Return(nil, sql.ErrNoRows)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockAudit.On("LogSystemAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	user, err := userService.CreateUser(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Admin", user.Role)
	assert.True(t, user.Active)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_UsernameExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	req := &models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Role:     "Admin",
	}

	existingUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "existing@example.com",
	}

	mockRepo.On("GetByUsername", "testuser").Return(existingUser, nil)

	user, err := userService.CreateUser(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "username 'testuser' already exists")
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	// Create a user with hashed password
	hashedPassword := "$2a$12$IvN20ZIyAUeHeR/TTal/RepPLN81XW6p.3Da//V6adrPSv.Y5MBUa" // "password123"
	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         "Admin",
		Active:       true,
	}

	req := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)
	mockAudit.On("LogUserAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	response, err := userService.Login(req, "127.0.0.1", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, user.ID, response.User.ID)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	req := &models.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", "nonexistent").Return(nil, sql.ErrNoRows)
	mockAudit.On("LogSystemAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	response, err := userService.Login(req, "127.0.0.1", "test-agent")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	existingUser := &models.User{
		ID:       1,
		Username: "olduser",
		Email:    "old@example.com",
		Role:     "PropertyManager",
		Active:   true,
	}

	req := &models.UpdateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Role:     "Admin",
	}

	mockRepo.On("GetByID", 1).Return(existingUser, nil)
	mockRepo.On("GetByUsername", "newuser").Return(nil, sql.ErrNoRows)
	mockRepo.On("GetByEmail", "new@example.com").Return(nil, sql.ErrNoRows)
	mockRepo.On("Update", 1, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockAudit.On("LogSystemAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := userService.UpdateUser(1, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	existingUser := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "Admin",
		Active:   true,
	}

	mockRepo.On("GetByID", 1).Return(existingUser, nil)
	mockRepo.On("Delete", 1).Return(nil)
	mockAudit.On("LogSystemAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := userService.DeleteUser(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	user := &models.User{
		ID:     1,
		Email:  "test@example.com",
		Active: true,
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
	mockRepo.On("CreateResetToken", mock.AnythingOfType("*models.ResetPasswordToken")).Return(nil)
	mockAudit.On("LogUserAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := userService.ResetPassword("test@example.com")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestConfirmPasswordReset_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	resetToken := &models.ResetPasswordToken{
		ID:        1,
		UserID:    1,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	user := &models.User{
		ID:     1,
		Active: true,
	}

	mockRepo.On("GetResetToken", "valid-token").Return(resetToken, nil)
	mockRepo.On("GetByID", 1).Return(user, nil)
	mockRepo.On("Update", 1, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRepo.On("MarkResetTokenUsed", 1).Return(nil)
	mockAudit.On("LogUserAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := userService.ConfirmPasswordReset("valid-token", "NewPassword123!")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChangePassword_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockAudit := new(MockAuditService)

	userService := &UserService{
		userRepo:      mockRepo,
		auditService:  mockAudit,
		jwtSecret:     "test-secret",
		jwtExpiration: 8 * time.Hour,
	}

	// Create a user with hashed password
	hashedPassword := "$2a$12$IvN20ZIyAUeHeR/TTal/RepPLN81XW6p.3Da//V6adrPSv.Y5MBUa" // "password123"
	user := &models.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}

	mockRepo.On("GetByID", 1).Return(user, nil)
	mockRepo.On("Update", 1, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockAudit.On("LogUserAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := userService.ChangePassword(1, "password123", "NewPassword123!")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
