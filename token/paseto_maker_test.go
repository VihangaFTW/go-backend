package token

import (
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(TestingHexKey)

	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(duration)

	// create the paseto token
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, payload.IssuedAt, issuedAt, time.Second)
	require.WithinDuration(t, payload.ExpiresAt, expiresAt, time.Second)
}

func TestExpiredPaseto(t *testing.T) {
	maker, err := NewPasetoMaker(TestingHexKey)

	require.NoError(t, err)

	username := util.RandomOwner()
	// create the paseto token
	token, err := maker.CreateToken(username, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestSKeyEnvNotSet(t *testing.T) {
	maker, err := NewPasetoMaker("") // simulate missing key
	require.Error(t, err)
	require.EqualError(t, err, ErrMissingPasetoEnvVariable.Error())
	require.Nil(t, maker)
}

func TestMalformedSKeyEnvVar(t *testing.T) {
	maker, err := NewPasetoMaker("43C235235A4B") // malformed key, shorter than 32 bytes
	require.Error(t, err)
	require.EqualError(t, err, ErrFailedSKeyConversion.Error())
	require.Nil(t, maker)
}

func TestInvalidKeyLength(t *testing.T) {
	shortKey := "1234567890abcdef" // only 16 bytes (32 hex chars), should be 32 bytes (64 hex chars)
	maker, err := NewPasetoMaker(shortKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key size")
	require.Nil(t, maker)
}
