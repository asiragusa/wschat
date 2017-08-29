package services

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TokenGeneratorTestSuite struct {
	suite.Suite
	TokenGenerator *Generator
	userRepository mocks.UserRepository
	keyFunc        jwt.Keyfunc

	signingKey string
	issuer     string
	audience   string
}

func (suite *TokenGeneratorTestSuite) SetupSuite() {
	suite.signingKey = "secret"
	suite.audience = "access"
	suite.issuer = "http://localhost"

	suite.TokenGenerator = NewTokenGenerator(suite.signingKey, suite.issuer, suite.audience)
	suite.keyFunc = func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	}
}

func (suite *TokenGeneratorTestSuite) SetupTest() {
	suite.TokenGenerator.Clock = clockwork.NewFakeClock()

	suite.userRepository = mocks.UserRepository{}
	suite.TokenGenerator.UserRepository = &suite.userRepository
}

func (suite *TokenGeneratorTestSuite) TestGenerator() {
	user := entity.User{
		Id:     "id",
		Email:  "test@test.com",
		Secret: "userSecret",
	}

	suite.TokenGenerator.Clock = clockwork.NewFakeClockAt(time.Now())

	signed, err := suite.TokenGenerator.GenerateToken(user, time.Hour*24)
	suite.Require().NoError(err)

	token, err := jwt.ParseWithClaims(signed, &jwt.StandardClaims{}, suite.keyFunc)
	suite.Require().NoError(err)

	claims, ok := token.Claims.(*jwt.StandardClaims)
	suite.Require().True(ok)

	suite.Equal(suite.audience, claims.Audience)
	suite.Equal(suite.issuer, claims.Issuer)
	suite.Equal(user.Id, claims.Subject)
	suite.Equal(user.Secret, claims.Id)

	now := suite.TokenGenerator.Clock.Now().Add(time.Hour * 24)
	suite.Equal(now.Unix(), claims.ExpiresAt)
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenWithInvalidToken() {
	token := "invalid"

	user, err := suite.TokenGenerator.ValidateToken(token)
	suite.Nil(user)
	suite.EqualError(err, InvalidTokenError.Error())
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenWithInvalidClaims() {
	invalidClaims := []jwt.StandardClaims{
		{
			Audience: "invalid",
			Issuer:   suite.issuer,
		},
		{
			Audience: suite.audience,
			Issuer:   "http://invalid",
		},
	}
	for _, claims := range invalidClaims {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(suite.signingKey))
		suite.Require().NoError(err)

		user, err := suite.TokenGenerator.ValidateToken(signed)
		suite.Nil(user)
		suite.EqualError(err, InvalidTokenError.Error())
	}
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenWithBadUserId() {
	claims := &jwt.StandardClaims{
		Audience: suite.audience,
		Issuer:   suite.issuer,
		Subject:  "invalid",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(suite.signingKey))
	suite.Require().NoError(err)

	suite.userRepository.On("GetUserById", claims.Subject).Return(nil, repository.UserNotFoundError)

	user, err := suite.TokenGenerator.ValidateToken(signed)
	suite.Nil(user)
	suite.EqualError(err, InvalidTokenError.Error())
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenWithRepoError() {
	claims := &jwt.StandardClaims{
		Audience: suite.audience,
		Issuer:   suite.issuer,
		Subject:  "userId",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(suite.signingKey))
	suite.Require().NoError(err)

	suite.userRepository.On("GetUserById", claims.Subject).Return(nil, assert.AnError)

	user, err := suite.TokenGenerator.ValidateToken(signed)
	suite.Nil(user)
	suite.EqualError(err, assert.AnError.Error())
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenWithBadSecret() {
	claims := jwt.StandardClaims{
		Audience: suite.audience,
		Issuer:   suite.issuer,
		Subject:  "userId",
		Id:       "badSecret",
	}
	mockUser := entity.User{
		Id:     claims.Subject,
		Secret: "secret",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signed, err := token.SignedString([]byte(suite.signingKey))
	suite.Require().NoError(err)

	suite.userRepository.On("GetUserById", mockUser.Id).Return(&mockUser, nil)

	user, err := suite.TokenGenerator.ValidateToken(signed)
	suite.Nil(user)
	suite.EqualError(err, InvalidTokenError.Error())
}

func (suite *TokenGeneratorTestSuite) TestValidateTokenOK() {
	claims := jwt.StandardClaims{
		Audience: suite.audience,
		Issuer:   suite.issuer,
		Subject:  "userId",
		Id:       "secret",
	}
	mockUser := entity.User{
		Id:     claims.Subject,
		Secret: claims.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(suite.signingKey))
	suite.Require().NoError(err)

	suite.userRepository.On("GetUserById", mockUser.Id).Return(&mockUser, nil)

	user, err := suite.TokenGenerator.ValidateToken(signed)
	suite.Require().NoError(err)
	suite.Require().NotNil(user)
	suite.Equal(user, &mockUser)
}

func TestConfirmTokenGenerator(t *testing.T) {
	suite.Run(t, new(TokenGeneratorTestSuite))
}
