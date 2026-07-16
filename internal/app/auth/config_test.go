package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/auth"
)

func TestConfigValidate(t *testing.T) {
	t.Run("rejects short JWT secret", func(t *testing.T) {
		c := auth.Config{JWTSecret: "too-short", SessionTTL: time.Hour}
		require.Error(t, c.Validate())
	})
	t.Run("accepts a strong secret", func(t *testing.T) {
		c := auth.Config{JWTSecret: "0123456789abcdef0123456789abcdef", SessionTTL: time.Hour}
		require.NoError(t, c.Validate())
	})
	t.Run("rejects non-positive TTL", func(t *testing.T) {
		c := auth.Config{JWTSecret: "0123456789abcdef0123456789abcdef", SessionTTL: 0}
		require.Error(t, c.Validate())
	})
}
