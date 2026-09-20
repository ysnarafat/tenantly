package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
)

// Role constants for authorization
const (
	RoleSuperAdmin      = "SUPER_ADMIN"
	RoleOrgAdmin        = "ORG_ADMIN"
	RoleAdmin           = "Admin"
	RolePropertyManager = "PropertyManager"
	RoleAccountant      = "Accountant"
)

// AuthMiddleware handles JWT authentication and sets user context
func AuthRequired(jwtSecret string, auditService *database.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			// Log failed authentication attempt
			if auditService != nil {
				_ = auditService.LogSystemAction(
					models.AuditActionLogin,
					models.TableUsers,
					nil,
					nil,
					map[string]interface{}{
						"error":      "invalid_token",
						"ip_address": c.ClientIP(),
						"user_agent": c.GetHeader("User-Agent"),
						"timestamp":  time.Now(),
					},
				)
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		// Extract and validate claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID in token",
				"code":  "INVALID_USER_ID",
			})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid role in token",
				"code":  "INVALID_ROLE",
			})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)

		// Set user context
		c.Set("user_id", int(userID))
		c.Set("role", role)
		c.Set("username", username)

		// Extract organization_id if present (optional for SUPER_ADMIN)
		var organizationID *int
		if orgIDClaim, ok := claims["organization_id"]; ok {
			if orgIDFloat, ok := orgIDClaim.(float64); ok {
				orgID := int(orgIDFloat)
				organizationID = &orgID
			}
		}
		c.Set("organization_id", organizationID)

		// Log successful authentication
		if auditService != nil {
			_ = auditService.LogUserAction(
				int(userID),
				"ACCESS",
				models.TableUsers,
				nil,
				nil,
				map[string]interface{}{
					"endpoint":   c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip_address": c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
					"timestamp":  time.Now(),
				},
			)
		}

		c.Next()
	}
}

// RequireRole creates middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid role type",
				"code":  "INVALID_ROLE_TYPE",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		// Note: unauthorized access attempts are audit-logged by the security
		// middleware layer, which has the auditService dependency injected.

		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("Access denied. Required roles: %v, user role: %s", roles, role),
			"code":  "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}

// RequireAdmin middleware for admin-only endpoints
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(RoleAdmin)
}

// RequireAdminOrPropertyManager middleware for admin or property manager endpoints
func RequireAdminOrPropertyManager() gin.HandlerFunc {
	return RequireRole(RoleAdmin, RolePropertyManager)
}

// RequireAdminOrPropertyManagerOrAccountant middleware for admin, property manager, or accountant endpoints
func RequireAdminOrPropertyManagerOrAccountant() gin.HandlerFunc {
	return RequireRole(RoleAdmin, RolePropertyManager, RoleAccountant)
}

// RequireSuperAdminOrAdmin allows SUPER_ADMIN, ORG_ADMIN, or Admin
func RequireSuperAdminOrAdmin() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin, RoleOrgAdmin, RoleAdmin)
}

// RequireAnyRole middleware that allows any authenticated user
func RequireAnyRole() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin, RoleOrgAdmin, RoleAdmin, RolePropertyManager, RoleAccountant)
}

// GetUserID helper function to extract user ID from context
func GetUserID(c *gin.Context) (int, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}

	id, ok := userID.(int)
	if !ok {
		return 0, fmt.Errorf("invalid user ID type")
	}

	return id, nil
}

// GetUserRole helper function to extract user role from context
func GetUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", fmt.Errorf("user role not found in context")
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", fmt.Errorf("invalid role type")
	}

	return roleStr, nil
}

// RequireOrgContext blocks requests where the JWT has no organization_id.
// Must run after AuthRequired. Injects "org_id" (int) into context for handlers.
func RequireOrgContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDRaw, exists := c.Get("organization_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "organization context required",
				"code":  "ORG_CONTEXT_REQUIRED",
			})
			c.Abort()
			return
		}
		orgIDPtr, ok := orgIDRaw.(*int)
		if !ok || orgIDPtr == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "organization context required",
				"code":  "ORG_CONTEXT_REQUIRED",
			})
			c.Abort()
			return
		}
		c.Set("org_id", *orgIDPtr)
		c.Next()
	}
}

// GetUsername helper function to extract username from context
func GetUsername(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("invalid username type")
	}

	return usernameStr, nil
}
