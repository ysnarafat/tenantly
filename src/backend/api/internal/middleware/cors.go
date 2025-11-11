package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple CORS headers - nginx proxy will handle the real CORS

		if environment != "development" {
			c.Next()
		}

		// Default fallback or keep as * if appropriate for dev, but specific origin is safer for credentials
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
