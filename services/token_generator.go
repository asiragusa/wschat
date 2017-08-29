package services

import (
	"crypto/subtle"
	"errors"
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/jonboulle/clockwork"
	"time"
)

var (
	// Error thrown when the token is invalid
	InvalidTokenError = errors.New("Invalid Token")
)

// Interface used mainly for Unit testing
type TokenGenerator interface {
	GenerateToken(entity.User, time.Duration) (string, error)
	ValidateToken(string) (*entity.User, error)
}

// JWT Token generator and validator
type Generator struct {
	// Injected via DI
	Clock clockwork.Clock `inject:""`

	// Injected via DI
	UserRepository repository.UserRepository `inject:""`

	signingKey []byte
	issuer     string
	audience   string
}

// Creates a new JWT Token Generator
// signingKey is the key used to sign the tokens
// issuer of the token eg. http://localhost
// audience for the token eg. access, ws etc
func NewTokenGenerator(signingKey, issuer, audience string) *Generator {
	return &Generator{
		signingKey: []byte(signingKey),
		issuer:     issuer,
		audience:   audience,
	}
}

// Generates a new token for the user with the duration duration
func (g Generator) GenerateToken(user entity.User, duration time.Duration) (string, error) {
	expiresAt := g.Clock.Now().Add(duration)

	claims := &jwt.StandardClaims{
		Subject:   user.Id,
		Id:        user.Secret,
		Audience:  g.audience,
		ExpiresAt: expiresAt.Unix(),
		Issuer:    g.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.signingKey)
}

// Validates the given signed string. Returns the user fetched from the DB if the token is valid
func (g Generator) ValidateToken(signed string) (*entity.User, error) {
	// Parse the signed string
	token, err := jwt.ParseWithClaims(signed, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return g.signingKey, nil
	})
	if err != nil {
		return nil, InvalidTokenError
	}

	// Get the claims from the token
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims.Audience != g.audience || claims.Issuer != g.issuer {
		return nil, InvalidTokenError
	}

	// Get the user from the DB
	user, err := g.UserRepository.GetUserById(claims.Subject)
	if err == repository.UserNotFoundError {
		return nil, InvalidTokenError
	}
	if err != nil {
		return nil, err
	}

	// Compare the user's secret with the given claim
	if subtle.ConstantTimeCompare([]byte(user.Secret), []byte(claims.Id)) != 1 {
		return nil, InvalidTokenError
	}

	return user, nil
}
