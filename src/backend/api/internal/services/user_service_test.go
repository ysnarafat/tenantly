package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ysnarafat/tenantly/internal/models"
)

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
