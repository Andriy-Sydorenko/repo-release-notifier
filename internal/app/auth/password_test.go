package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/app/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct horse battery staple")
	require.NoError(t, err)
	require.Contains(t, hash, "$argon2id$")

	ok, err := auth.VerifyPassword(hash, "correct horse battery staple")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = auth.VerifyPassword(hash, "wrong password")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestHashPasswordProducesDistinctHashes(t *testing.T) {
	h1, err := auth.HashPassword("same-password")
	require.NoError(t, err)
	h2, err := auth.HashPassword("same-password")
	require.NoError(t, err)
	require.NotEqual(t, h1, h2, "per-hash random salt must make hashes differ")
}

func TestVerifyPasswordRejectsGarbage(t *testing.T) {
	_, err := auth.VerifyPassword("not-a-phc-string", "x")
	require.ErrorIs(t, err, auth.ErrInvalidHash)
}
