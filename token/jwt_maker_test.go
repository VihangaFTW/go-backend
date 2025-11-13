package token

import (
	"errors"
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// TestJWTMaker tests the happy path of JWT token creation and verification
// This ensures our JWT implementation works correctly under normal conditions
func TestJWTMaker(t *testing.T) {
	// Step 1: Create a JWT maker with a random 32-character secret key
	// We use a random key to ensure each test run is independent
	// The secret key is used to sign and verify JWT tokens
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err) // Ensure JWT maker creation doesn't fail

	// Step 2: Define test data for creating a token
	// We use random data to avoid test dependencies and ensure uniqueness
	username := util.RandomOwner()      // Random username for the token payload
	duration := time.Minute             // Token will be valid for 1 minute
	issuedAt := time.Now()              // Current time as issue time
	expiresAt := issuedAt.Add(duration) // Calculate expiration time

	// Step 3: Create a JWT token using our maker
	// This tests the CreateToken functionality
	token, payload, err := maker.CreateToken(username, duration)
	require.NoError(t, err)    // Token creation should succeed
	require.NotEmpty(t, token) // Token should not be empty string
	require.NotEmpty(t, payload)

	// Step 4: Verify the token and extract its payload
	// This tests the VerifyToken functionality and ensures the round-trip works
	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)      // Token verification should succeed
	require.NotEmpty(t, payload) // Payload should not be nil

	// Step 5: Validate all payload fields are correct
	// This ensures the token contains exactly what we put into it
	require.NotZero(t, payload.ID)                                       // ID should be generated (UUID)
	require.Equal(t, payload.Username, username)                         // Username should match input
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)   // IssuedAt should be close to our timestamp
	require.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Second) // ExpiresAt should be close to our calculated time
}

// TestExpiredJWTToken tests that expired tokens are properly rejected
// This is crucial for security - expired tokens should never be accepted
func TestExpiredJWTToken(t *testing.T) {
	// Step 1: Create a JWT maker with a random secret key
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)    // JWT maker creation should succeed
	require.NotEmpty(t, maker) // Maker should not be nil

	// Step 2: Prepare test data
	username := util.RandomOwner()

	// Step 3: Create a token that's already expired
	// We use -time.Minute to create a token that expired 1 minute ago
	// This simulates a real-world scenario where a token has expired
	token, payload, err := maker.CreateToken(username, -time.Minute)
	require.NoError(t, err)    // Token creation should still succeed
	require.NotEmpty(t, token) // Token should be created (expiration is checked during verification)
	require.NotEmpty(t, payload)

	// Step 4: Try to verify the expired token
	// This should fail because the token is expired
	payload, err = maker.VerifyToken(token)
	require.Error(t, err) // Verification should fail

	// Step 5: Ensure we get the correct error type
	// The JWT v5 package wraps errors for better debugging context
	// We need to check the underlying error to ensure it's specifically an expiration error
	require.True(t, errors.Is(err, jwt.ErrTokenExpired))

	// Step 6: Ensure no payload is returned for invalid tokens
	// This prevents accidental use of data from invalid tokens
	require.Nil(t, payload)
}

// TestInvalidJWTTokenAlgNone tests protection against the "none" algorithm attack
// This is a critical security test - tokens signed with "none" algorithm should be rejected
// The "none" algorithm attack allows attackers to create unsigned tokens that might be accepted
func TestInvalidJWTTokenAlgNone(t *testing.T) {
	// Step 1: Create a valid payload for testing
	// We create a legitimate payload to ensure the rejection is due to the algorithm, not the content
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err) // Payload creation should succeed

	// Step 2: Convert payload to JWT claims format
	jwtClaims := NewJWTPayloadClaims(payload)

	// Step 3: Create a malicious token using the "none" algorithm
	// This simulates an attacker trying to create an unsigned token
	// The "none" algorithm means the token has no signature
	tokenStruct := jwt.NewWithClaims(jwt.SigningMethodNone, jwtClaims)
	token, err := tokenStruct.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err) // Token creation should succeed (we're testing the attack scenario)

	// Step 4: Create a legitimate JWT maker that should reject the malicious token
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err) // JWT maker creation should succeed

	// Step 5: Try to verify the malicious token
	// Our JWT maker should reject tokens with "none" algorithm
	payload, err = maker.VerifyToken(token)
	require.Error(t, err) // Verification should fail

	// Step 6: Ensure we get the correct error type
	// Should specifically be a signature validation error
	require.True(t, errors.Is(err, jwt.ErrTokenSignatureInvalid))

	// Step 7: Ensure no payload is returned for the malicious token
	// This prevents any accidental use of the unsigned token's data
	require.Nil(t, payload)
}
