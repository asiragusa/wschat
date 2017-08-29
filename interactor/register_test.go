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

type RegisterInteractorTestSuite struct {
	suite.Suite
	interactor     *Register
	userRepository *mocks.UserRepository

	accessTokenGenerator *mocks.TokenGenerator
}

func TestRegisterInteractor(t *testing.T) {
	suite.Run(t, new(RegisterInteractorTestSuite))
}

func (suite *RegisterInteractorTestSuite) SetupSuite() {
	suite.interactor = NewRegisterInteractor()
}

func (suite *RegisterInteractorTestSuite) SetupTest() {
	suite.userRepository = &mocks.UserRepository{}

	suite.accessTokenGenerator = &mocks.TokenGenerator{}

	suite.interactor.UserRepository = suite.userRepository
	suite.interactor.AccessTokenGenerator = suite.accessTokenGenerator
}

func (suite *RegisterInteractorTestSuite) TearDownTest() {
	suite.userRepository.AssertExpectations(suite.T())
	suite.accessTokenGenerator.AssertExpectations(suite.T())
}

func (suite *RegisterInteractorTestSuite) getValidRequest() request.Register {
	return request.Register{
		Email:    "a@b.com",
		Password: "validPassword",
	}
}

func (suite *RegisterInteractorTestSuite) TestUsernameAlreadyExists() {
	request := suite.getValidRequest()

	suite.userRepository.On("CreateUser", request.Email, request.Password).Return(nil, repository.UserAlreadyExistsError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.NewError(httptest.StatusUnprocessableEntity)
	expected.AddDetail("email", "alreadyExists")

	suite.Equal(expected, r)
}

func (suite *RegisterInteractorTestSuite) TestRepositoryAnyError() {
	request := suite.getValidRequest()

	suite.userRepository.On("CreateUser", request.Email, request.Password).Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *RegisterInteractorTestSuite) TestAccessTokenGeneratorAnyError() {
	request := suite.getValidRequest()
	user := entity.User{Email: request.Email}

	suite.userRepository.On("CreateUser", request.Email, request.Password).Return(&user, nil)
	suite.accessTokenGenerator.On("GenerateToken", user, time.Hour*24*30).Return("", assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *RegisterInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()
	user := entity.User{Email: request.Email}

	suite.userRepository.On("CreateUser", request.Email, request.Password).Return(&user, nil)
	suite.accessTokenGenerator.On("GenerateToken", user, time.Hour*24*30).Return("access", nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.Register{
		AccessToken: "access",
	}
	suite.Equal(expected, r)
}
