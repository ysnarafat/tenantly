package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ysnarafat/tenantly/internal/services"
)

// StepUpHeader is the request header carrying the short-lived MFA step-up token.
const StepUpHeader = "X-Step-Up-Token"

// RequireStepUp gates an endpoint behind a recent MFA verification. The caller
// must present a valid, unexpired step-up token (from POST /auth/mfa/verify)
// belonging to the authenticated user. Must run after AuthRequired.
func RequireStepUp(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader(StepUpHeader)
		if tokenString == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "MFA verification required",
				"code":  "MFA_STEP_UP_REQUIRED",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid or expired MFA verification",
				"code":  "MFA_STEP_UP_INVALID",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["purpose"] != services.StepUpPurpose {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid MFA verification token",
				"code":  "MFA_STEP_UP_INVALID",
			})
			c.Abort()
			return
		}

		// The step-up token must belong to the authenticated caller.
		uid, ok := claims["user_id"].(float64)
		if !ok || int(uid) != c.GetInt("user_id") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "MFA verification does not match the authenticated user",
				"code":  "MFA_STEP_UP_INVALID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
