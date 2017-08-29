package interactor

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type LoginInteractorTestSuite struct {
	suite.Suite
	interactor     *Login
	userRepository *mocks.UserRepository

	accessTokenGenerator *mocks.TokenGenerator
}

func TestLoginInteractor(t *testing.T) {
	suite.Run(t, new(LoginInteractorTestSuite))
}

func (suite *LoginInteractorTestSuite) SetupSuite() {
	suite.interactor = NewLoginInteractor()
}

func (suite *LoginInteractorTestSuite) SetupTest() {
	suite.userRepository = &mocks.UserRepository{}

	suite.accessTokenGenerator = &mocks.TokenGenerator{}

	suite.interactor.UserRepository = suite.userRepository
	suite.interactor.AccessTokenGenerator = suite.accessTokenGenerator
}

func (suite *LoginInteractorTestSuite) TearDownTest() {
	suite.userRepository.AssertExpectations(suite.T())
	suite.accessTokenGenerator.AssertExpectations(suite.T())
}

func (suite *LoginInteractorTestSuite) getValidRequest() request.Login {
	return request.Login{
		Email:    "a@b.com",
		Password: "validPassword",
	}
}

func (suite *LoginInteractorTestSuite) TestRepositoryBadUsernameOrPassword() {
	request := suite.getValidRequest()

	suite.userRepository.On("Login", request.Email, request.Password).Return(nil, repository.UserBadUsernameOrPasswordError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusUnauthorized), r)
}

func (suite *LoginInteractorTestSuite) TestRepositoryAnyError() {
	request := suite.getValidRequest()

	suite.userRepository.On("Login", request.Email, request.Password).Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *LoginInteractorTestSuite) TestAccessTokenGeneratorAnyError() {
	request := suite.getValidRequest()
	user := entity.User{Email: request.Email}

	suite.userRepository.On("Login", request.Email, request.Password).Return(&user, nil)
	suite.accessTokenGenerator.On("GenerateToken", user, time.Hour*24*30).Return("", assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *LoginInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()
	user := entity.User{Email: request.Email}

	suite.userRepository.On("Login", request.Email, request.Password).Return(&user, nil)
	suite.accessTokenGenerator.On("GenerateToken", user, time.Hour*24*30).Return("access", nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.Login{
		AccessToken: "access",
	}
	suite.Equal(expected, r)
}
