package token

import (
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/google/uuid"
)

// PasetoMaker is a struct that holds the PASETO symmetric key for token operations.
type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
	implicit     []byte
}

// NewPasetoMaker creates a new PasetoMaker instance with the provided symmetric key.
// It returns an error if the key is not valid.
func NewPasetoMaker(config util.Config) (Maker, error) {

	key := config.PasetoHexKey

	//? first make sure we have a secret key to encrypt payload
	if key == "" {
		return nil, ErrMissingPasetoEnvVariable
	}

	symmetricKey, err := paseto.V4SymmetricKeyFromHex(key)
	if err != nil {
		return nil, ErrFailedSKeyConversion
	}

	return &PasetoMaker{
		symmetricKey, []byte{},
	}, nil
}

func (p *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// create the paseto token
	token := paseto.NewToken()

	// create uuid for token id
	tokenId, err := uuid.NewRandom()

	if err != nil {
		return "", err
	}

	// add data to token
	token.Set("id", tokenId)
	token.Set("username", username)
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(duration))

	return token.V4Encrypt(p.symmetricKey, p.implicit), nil

}

func (p *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())
	parsedToken, err := parser.ParseV4Local(p.symmetricKey, token, p.implicit)

	if err != nil {
		return nil, ErrExpiredToken
	}

	payload, err := getPayloadFromToken(parsedToken)

	if err != nil {
		return nil, err
	}

	return payload, nil
}

func getPayloadFromToken(t *paseto.Token) (*Payload, error) {
	id, err := t.GetString("id")

	if err != nil {
		return nil, ErrInvalidToken
	}

	username, err := t.GetString("username")

	if err != nil {
		return nil, ErrInvalidToken
	}

	issuedAt, err := t.GetIssuedAt()

	if err != nil {
		return nil, ErrInvalidToken
	}

	expiresAt, err := t.GetExpiration()

	if err != nil {
		return nil, ErrInvalidToken
	}

	return &Payload{
		ID:        uuid.MustParse(id),
		Username:  username,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil

}
