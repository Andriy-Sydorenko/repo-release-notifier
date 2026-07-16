package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/api"
	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/auth"
)

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	r := gin.New()
	r.GET("/me", api.RequireAuth(iss), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": c.GetUint("userID")})
	})

	t.Run("401 without cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", http.NoBody))
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("401 with garbage cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", http.NoBody)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: "not-a-jwt"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("200 with valid cookie", func(t *testing.T) {
		tok, _ := iss.Issue(99, "u@example.com", 0)
		req := httptest.NewRequest(http.MethodGet, "/me", http.NoBody)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: tok})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "99")
	})
}
