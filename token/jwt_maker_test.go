package token

import (
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/db/util"
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

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, payload.Username, username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Second)

}
