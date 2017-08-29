package interactor

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ListUsersInteractorTestSuite struct {
	suite.Suite
	interactor     *ListUsers
	userRepository *mocks.UserRepository
}

func TestListUsersInteractor(t *testing.T) {
	suite.Run(t, new(ListUsersInteractorTestSuite))
}

func (suite *ListUsersInteractorTestSuite) SetupSuite() {
	suite.interactor = NewListUsersInteractor()
}

func (suite *ListUsersInteractorTestSuite) SetupTest() {
	suite.userRepository = &mocks.UserRepository{}

	suite.interactor.UserRepository = suite.userRepository
}

func (suite *ListUsersInteractorTestSuite) TearDownTest() {
	suite.userRepository.AssertExpectations(suite.T())
}

func (suite *ListUsersInteractorTestSuite) getValidRequest() request.ListUsers {
	return request.ListUsers{}
}

func (suite *ListUsersInteractorTestSuite) TestRepositoryAnyError() {
	request := suite.getValidRequest()

	suite.userRepository.On("All").Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *ListUsersInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()
	users := []entity.User{
		{Email: "a@b.com"},
		{Email: "b@b.com"},
	}

	suite.userRepository.On("All").Return(users, nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.ListUsers{
		Total: 2,
		Items: []response.User{
			{Email: "a@b.com"},
			{Email: "b@b.com"},
		},
	}
	suite.Equal(expected, r)
}
