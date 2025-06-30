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
	token,payload,err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err = maker.VerifyToken(token)
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
	token, payload, err := maker.CreateToken(username, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token
	payload, err = maker.VerifyToken(token)
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

func TestShortHexKey(t *testing.T) {
	maker, err := NewPasetoMaker("23432423ab32cd") //  shorter than 64 hex characters
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidKeySize.Error())
	require.Nil(t, maker)
}

func TestInvalidHexKey(t *testing.T) {
	// this key contains a "/" at the end which is not a part of hex encoding 
	notHexKey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde/"
	maker, err := NewPasetoMaker(notHexKey)
	require.Error(t, err)
	require.EqualError(t, err, ErrFailedSKeyConversion.Error())
	require.Nil(t, maker)
}
