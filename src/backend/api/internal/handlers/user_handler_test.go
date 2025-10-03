package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/models"
)

// MockUserService is a mock implementation of UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(req *models.LoginRequest, clientIP, userAgent string) (*models.LoginResponse, error) {
	args := m.Called(req, clientIP, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockUserService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockUserService) Logout(userID int, clientIP, userAgent string) error {
	args := m.Called(userID, clientIP, userAgent)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(userID int, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) ConfirmPasswordReset(token, newPassword string) error {
	args := m.Called(token, newPassword)
	return args.Error(0)
}

func (m *MockUserService) GetUserByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id int, req *models.UpdateUserRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.POST("/users", handler.CreateUser)

	t.Run("Success", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "ValidPass123!",
			Role:     "Admin",
		}

		user := &models.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "Admin",
			Active:   true,
		}

		mockService.On("CreateUser", req).Return(user, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, response.Username)
		assert.Equal(t, user.Email, response.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_Login(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.POST("/auth/login", handler.Login)

	t.Run("Success", func(t *testing.T) {
		req := &models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		loginResponse := &models.LoginResponse{
			Token:        "jwt-token",
			RefreshToken: "refresh-token",
			User: models.User{
				ID:       1,
				Username: "testuser",
				Role:     "Admin",
			},
			ExpiresAt: time.Now().Add(8 * time.Hour),
		}

		mockService.On("Login", req, "", "").Return(loginResponse, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, loginResponse.Token, response.Token)
		assert.Equal(t, loginResponse.User.Username, response.User.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		req := &models.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		mockService.On("Login", req, "", "").Return(nil, assert.AnError)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_RefreshToken(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.POST("/auth/refresh", handler.RefreshToken)

	t.Run("Success", func(t *testing.T) {
		req := &models.RefreshTokenRequest{
			RefreshToken: "valid-refresh-token",
		}

		loginResponse := &models.LoginResponse{
			Token:        "new-jwt-token",
			RefreshToken: "new-refresh-token",
			User: models.User{
				ID:       1,
				Username: "testuser",
				Role:     "Admin",
			},
			ExpiresAt: time.Now().Add(8 * time.Hour),
		}

		mockService.On("RefreshToken", "valid-refresh-token").Return(loginResponse, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, loginResponse.Token, response.Token)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	// Middleware to set user context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})

	router.POST("/auth/change-password", handler.ChangePassword)

	t.Run("Success", func(t *testing.T) {
		req := &models.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "NewPassword123!",
		}

		mockService.On("ChangePassword", 1, "oldpassword", "NewPassword123!").Return(nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_ResetPassword(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.POST("/auth/reset-password", handler.ResetPassword)

	t.Run("Success", func(t *testing.T) {
		req := &models.ResetPasswordRequest{
			Email: "test@example.com",
		}

		mockService.On("ResetPassword", "test@example.com").Return(nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_ConfirmPasswordReset(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.POST("/auth/confirm-reset-password", handler.ConfirmPasswordReset)

	t.Run("Success", func(t *testing.T) {
		req := &models.ConfirmPasswordResetRequest{
			Token:       "valid-token",
			NewPassword: "NewPassword123!",
		}

		mockService.On("ConfirmPasswordReset", "valid-token", "NewPassword123!").Return(nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("POST", "/auth/confirm-reset-password", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUsers(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.GET("/users", handler.GetUsers)

	t.Run("Success", func(t *testing.T) {
		users := []*models.User{
			{
				ID:       1,
				Username: "user1",
				Email:    "user1@example.com",
				Role:     "Admin",
				Active:   true,
			},
			{
				ID:       2,
				Username: "user2",
				Email:    "user2@example.com",
				Role:     "PropertyManager",
				Active:   true,
			},
		}

		mockService.On("GetAllUsers").Return(users, nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("GET", "/users", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string][]*models.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response["users"], 2)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.PUT("/users/:id", handler.UpdateUser)

	t.Run("Success", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			Username: "updateduser",
			Email:    "updated@example.com",
			Role:     "PropertyManager",
		}

		mockService.On("UpdateUser", 1, req).Return(nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("PUT", "/users/invalid", bytes.NewBuffer([]byte("{}")))
		httpReq.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	router.DELETE("/users/:id", handler.DeleteUser)

	t.Run("Success", func(t *testing.T) {
		mockService.On("DeleteUser", 1).Return(nil)

		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/users/1", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		httpReq, _ := http.NewRequest("DELETE", "/users/invalid", nil)

		router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
