package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	minSecretKeySize = 32
	issuer           = "simplebank"
)

// JWTMaker is a JSON Web Token maker. It implements the Maker interface.
type JWTMaker struct {
	secretKey string
}

// JWTPayloadClaims is a wrapper that adds the registered claims.
type JWTPayloadClaims struct {
	Payload
	jwt.RegisteredClaims
}

func NewJWTPayloadClaims(payload *Payload) *JWTPayloadClaims {
	return &JWTPayloadClaims{
		Payload: *payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(payload.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			NotBefore: jwt.NewNumericDate(payload.IssuedAt),
			Issuer:    issuer,
			Subject:   payload.Username,
			ID:        payload.ID.String(),
			Audience:  []string{"clients"},
		},
	}
}

// NewJWTMaker creates a new JWTMaker object.
// The function returns a Maker interface as a safety check to ensure all required
// methods are implemented.
func NewJWTMaker(secretKey string) (Maker, error) {

	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

func (m *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)

	if err != nil {
		return "", err
	}
	//* create an unsigned JWT token struct
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, NewJWTPayloadClaims(payload))
	//* serialize the struct into the standard JWT string format (header.payload.signature)
	signedString, err := jwtToken.SignedString([]byte(m.secretKey))
	return signedString, err
}

func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {

	// keyFunc is a callback function that jwt libary calls to verify the signing algorithm.
	// Returns the secret key for signature verification.
	keyFunc := func(token *jwt.Token) (any, error) {
		// assert whether the signing method field in the given token is the expected algorithm
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		//? algorithm on the token does not match the server's signing algorithm
		if !ok {
			return nil, jwt.ErrTokenSignatureInvalid

		}

		// token signing method verified, return server's secret key to verify contents of jwt
		return []byte(m.secretKey), nil
	}

	// parses the JWT string into its components (header.payload.signature).
	// verifies the signature using the secret key from keyFunc
	// validates expiration and other standard claims
	// umarshals the payload into JWTPayloadClaims struct
	jwtToken, err := jwt.ParseWithClaims(token, &JWTPayloadClaims{}, keyFunc)

	if err != nil {
		return nil, err
	}

	// type assert to ensure the claims are of the expected type.
	// This is a safety check to make sure the parsing worked correctly
	payloadClaims, ok := jwtToken.Claims.(*JWTPayloadClaims)

	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	// return the extracted Payload from the verified token
	return &payloadClaims.Payload, nil

}
