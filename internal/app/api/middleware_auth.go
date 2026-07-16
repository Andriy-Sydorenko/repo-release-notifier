package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/auth"
)

// RequireAuth aborts with 401 unless the request carries a valid session cookie.
// On success it stores userID (uint) and userEmail (string) on the context.
func RequireAuth(issuer *auth.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !parseSession(c, issuer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Next()
	}
}

// OptionalAuth populates userID/userEmail when a valid session cookie is present
// and always continues, for pages that render differently when signed in.
func OptionalAuth(issuer *auth.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		parseSession(c, issuer)
		c.Next()
	}
}

// parseSession validates the session cookie and, on success, stores userID/userEmail
// on the context. It reports whether a valid session was found.
func parseSession(c *gin.Context, issuer *auth.Issuer) bool {
	raw, err := c.Cookie(auth.SessionCookie)
	if err != nil || raw == "" {
		return false
	}
	claims, err := issuer.Parse(raw)
	if err != nil {
		return false
	}
	uid, err := claims.UserID()
	if err != nil {
		return false
	}
	c.Set("userID", uid)
	c.Set("userEmail", claims.Email)
	return true
}
