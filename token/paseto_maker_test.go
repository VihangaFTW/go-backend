package token

import (
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/stretchr/testify/require"
)

var testingHexKey = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

func TestPasetoMaker(t *testing.T) {
	// mock env variable
	config := util.Config{PasetoHexKey: testingHexKey}
	maker, err := NewPasetoMaker(config)

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
	// mock env variable
	config := util.Config{PasetoHexKey: testingHexKey}
	maker, err := NewPasetoMaker(config)

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
	config := &util.Config{
		PasetoHexKey: "", // simulate missing key
	}

	maker, err := NewPasetoMaker(*config)
	require.Error(t, err)
	require.EqualError(t, err, ErrMissingPasetoEnvVariable.Error())
	require.Nil(t, maker)
}
