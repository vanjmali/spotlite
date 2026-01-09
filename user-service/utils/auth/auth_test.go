package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPasswordAndCompare(t *testing.T) {
	pw := "StrongPass123!"
	hash, err := HashPassword(pw)
	require.NoError(t, err)
	require.NotEqual(t, pw, hash)

	require.NoError(t, CompareHashAndPassword(hash, pw))
	require.Error(t, CompareHashAndPassword(hash, "WrongPass123!"))
}

func TestGenerateOTP(t *testing.T) {
	otp, err := GenerateOTP()
	require.NoError(t, err)
	require.Len(t, otp, 6)

	for i := range len(otp) {
		require.True(t, otp[i] >= '0' && otp[i] <= '9', "otp should be numeric")
	}
}

func TestRefreshTokenGenerationAndHash(t *testing.T) {
	token, err := GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)

	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)
	require.Equal(t, hash1, hash2)

	other, err := GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, other)
	require.NotEqual(t, hash1, HashRefreshToken(other))
}
