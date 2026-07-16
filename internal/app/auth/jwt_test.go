package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/auth"
)

func TestIssueAndParse(t *testing.T) {
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	tok, err := iss.Issue(42, "a@example.com", 7)
	require.NoError(t, err)

	claims, err := iss.Parse(tok)
	require.NoError(t, err)
	require.Equal(t, "a@example.com", claims.Email)
	require.Equal(t, 7, claims.TokenVersion)
	uid, err := claims.UserID()
	require.NoError(t, err)
	require.Equal(t, uint(42), uid)
}

func TestParseRejectsWrongSecret(t *testing.T) {
	tok, _ := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour).Issue(1, "a@b.c", 0)
	_, err := auth.NewIssuer([]byte("ffffffffffffffffffffffffffffffff"), time.Hour).Parse(tok)
	require.Error(t, err)
}

func TestParseRejectsExpired(t *testing.T) {
	tok, _ := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), -time.Minute).Issue(1, "a@b.c", 0)
	_, err := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour).Parse(tok)
	require.Error(t, err)
}
