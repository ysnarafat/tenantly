package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

// respondError logs the real error server-side and writes only a generic,
// safe message to the client — raw Go/DB error strings (table/column names,
// SQL fragments, file paths) must never reach API responses (OWASP A09:
// information disclosure via error handling).
func respondError(c *gin.Context, status int, code, publicMsg string, err error) {
	if err != nil {
		log.Printf("[%s] %s %s -> %d %s: %v", code, c.Request.Method, c.Request.URL.Path, status, publicMsg, err)
	}
	c.JSON(status, gin.H{"error": publicMsg, "code": code})
}
