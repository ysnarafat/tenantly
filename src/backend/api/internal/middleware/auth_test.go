package middleware

import (
	"encoding/json"
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

	// Create a valid test token
	claims := jwt.MapClaims{
		"user_id":  1,
		"username": "testuser",
		"role":     "Admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	assert.NoError(t, err)

	// Create an expired token
	expiredClaims := jwt.MapClaims{
		"user_id":  1,
		"username": "testuser",
		"role":     "Admin",
		"exp":      time.Now().Add(-time.Hour).Unix(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(jwtSecret))

	// Create token with wrong secret
	wrongSecretToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongSecretTokenString, _ := wrongSecretToken.SignedString([]byte("wrong-secret"))

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenString,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
		},
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "AUTH_HEADER_MISSING",
		},
		{
			name:           "Invalid token format (missing Bearer)",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_TOKEN_FORMAT",
		},
		{
			name:           "Malformed JWT",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_INVALID",
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer " + expiredTokenString,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_INVALID",
		},
		{
			name:           "Token signed with wrong secret",
			authHeader:     "Bearer " + wrongSecretTokenString,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_INVALID",
		},
		{
			name:           "Empty Bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthRequired(jwtSecret, nil))
			router.GET("/test", func(c *gin.Context) {
				// Verify context is populated with user info
				userID, exists := c.Get("user_id")
				if exists {
					assert.NotNil(t, userID)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, response["code"])
			}
		})
	}
}

func TestAuthRequiredContextPopulation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSecret := "test-secret"

	claims := jwt.MapClaims{
		"user_id":  42,
		"username": "testuser",
		"role":     "Admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	t.Run("Valid token populates all context fields", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthRequired(jwtSecret, nil))
		router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			assert.True(t, exists, "user_id should exist in context")

			// user_id is stored as int in context (converted from float64 by middleware)
			userIDVal, ok := userID.(int)
			assert.True(t, ok, "user_id should be int")
			assert.Equal(t, 42, userIDVal)

			username, exists := c.Get("username")
			assert.True(t, exists, "username should exist in context")
			assert.Equal(t, "testuser", username)

			role, exists := c.Get("role")
			assert.True(t, exists, "role should exist in context")
			assert.Equal(t, "Admin", role)

			c.JSON(http.StatusOK, gin.H{
				"message": "success",
			})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
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
