package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const SessionCookie = "session"

// SetSession writes the session token as an httpOnly, SameSite=Lax cookie.
// secure should be true in production (HTTPS).
func SetSession(c *gin.Context, token string, ttl time.Duration, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, token, int(ttl.Seconds()), "/", "", secure, true)
}

// ClearSession expires the session cookie.
func ClearSession(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, "", -1, "/", "", secure, true)
}
