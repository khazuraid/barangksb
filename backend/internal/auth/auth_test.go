package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTRoundTripAndValidation(t *testing.T) {
	secret := "super-secret-test-key-32-bytes!!"
	tok, err := GenerateToken(secret, "u-1", "admin@kantor.id", "admin", "Admin", 3, time.Hour)
	require.NoError(t, err)

	claims, err := ParseToken(secret, tok)
	require.NoError(t, err)
	assert.Equal(t, "u-1", claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, 3, claims.TokenVersion)

	_, err = ParseToken("wrong-secret", tok)
	assert.Error(t, err)

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: "u-1",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	})
	expiredStr, err := expired.SignedString([]byte(secret))
	require.NoError(t, err)
	_, err = ParseToken(secret, expiredStr)
	assert.Error(t, err)

	noneTok, err := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: "u-1", Role: "admin"}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	_, err = ParseToken(secret, noneTok)
	assert.Error(t, err)
}

func TestRBACEnforcerMatrix(t *testing.T) {
	e, err := NewEnforcer()
	require.NoError(t, err)

	allowed, err := e.Enforce("admin", "/api/users", "DELETE")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = e.Enforce("petugas", "/api/users", "GET")
	require.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = e.Enforce("petugas", "/api/items", "POST")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = e.Enforce("viewer", "/api/items", "POST")
	require.NoError(t, err)
	assert.False(t, allowed)
}
