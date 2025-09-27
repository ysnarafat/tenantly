package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSecret := "test-secret"

	// Create a test token
	claims := jwt.MapClaims{
		"user_id":  1,
		"username": "testuser",
		"role":     "Admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthRequired(jwtSecret, nil))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       string
		requiredRoles  []string
		expectedStatus int
	}{
		{
			name:           "Admin accessing admin endpoint",
			userRole:       RoleAdmin,
			requiredRoles:  []string{RoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PropertyManager accessing admin endpoint",
			userRole:       RolePropertyManager,
			requiredRoles:  []string{RoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "PropertyManager accessing allowed endpoint",
			userRole:       RolePropertyManager,
			requiredRoles:  []string{RoleAdmin, RolePropertyManager},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Accountant accessing read-only endpoint",
			userRole:       RoleAccountant,
			requiredRoles:  []string{RoleAdmin, RolePropertyManager, RoleAccountant},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Mock middleware to set user context
			router.Use(func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("role", tt.userRole)
				c.Set("username", "testuser")
				c.Next()
			})

			router.Use(RequireRole(tt.requiredRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
