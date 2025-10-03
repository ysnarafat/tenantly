package interfaces

import "github.com/ysnarafat/tenantly/internal/models"

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id int) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(id int, updates map[string]interface{}) error
	Delete(id int) error
	CreateResetToken(token *models.ResetPasswordToken) error
	GetResetToken(token string) (*models.ResetPasswordToken, error)
	MarkResetTokenUsed(tokenID int) error
	CleanupExpiredTokens() error
}

// AuditServiceInterface defines the interface for audit service operations
type AuditServiceInterface interface {
	LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error
	LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error
}

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	CreateUser(req *models.CreateUserRequest) (*models.User, error)
	Login(req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error)
	RefreshToken(refreshToken string) (*models.LoginResponse, error)
	Logout(userID int, clientIP, userAgent string) error
	ChangePassword(userID int, currentPassword, newPassword string) error
	ResetPassword(email string) error
	ConfirmPasswordReset(token, newPassword string) error
	GetUserByID(id int) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(id int, req *models.UpdateUserRequest) error
	DeleteUser(id int) error
}
