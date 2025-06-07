package token

import (
	"errors"
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	// make a jwt with random secret key
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	// define fields for test payload
	username := util.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(duration)

	// create a jwt token using the test payload
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify the integrity of the token and its payload contents
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, payload.Username, username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Second)

}

func TestExpiredJWTToken(t *testing.T) {
	// make a jwt with a random secret key
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, maker)

	username := util.RandomOwner()

	token, err := maker.CreateToken(username, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// verify token validity
	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	// v5 jwt package wraps errors in their new validation system for better debug context
	// check the underlying error
	require.True(t, errors.Is(err, jwt.ErrTokenExpired))


	require.Nil(t, payload)
}
