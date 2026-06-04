package jwt

import (
	"testing"
	"time"

	jwtLib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidate(t *testing.T) {
	userID := uuid.New().String()
	secretKey := "testing-sercret"
	duration := time.Hour

	tokenStr, err := GenerateToken(userID, duration, secretKey)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	payload, err := ValidateToken(tokenStr, secretKey)
	require.NoError(t, err)
	require.NotNil(t, payload)

	assert.Equal(t, userID, payload.UserID)
}

func TestErrorTypes(t *testing.T) {
	userID := uuid.New().String()
	secretKey := "testing-secret"

	assert.NotNil(t, ErrInvalidToken)
	assert.NotNil(t, ErrTokenCreation)
	assert.NotNil(t, ErrNoUserID)
	assert.NotNil(t, ErrBadAlgorithm)

	token := jwtLib.NewWithClaims(jwtLib.SigningMethodRS256, jwtLib.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SigningString()
	require.NoError(t, err)

	_, err = ValidateToken(tokenStr, secretKey)
	assert.Error(t, err)
	if err != nil {
		t.Logf("Error type: %v", err)
	}
}
