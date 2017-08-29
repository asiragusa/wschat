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

type ListMessagesInteractorTestSuite struct {
	suite.Suite
	interactor        *ListMessages
	messageRepository *mocks.MessageRepository
}

func TestListMessagesInteractor(t *testing.T) {
	suite.Run(t, new(ListMessagesInteractorTestSuite))
}

func (suite *ListMessagesInteractorTestSuite) SetupSuite() {
	suite.interactor = NewListMessagesInteractor()
}

func (suite *ListMessagesInteractorTestSuite) SetupTest() {
	suite.messageRepository = &mocks.MessageRepository{}

	suite.interactor.MessageRepository = suite.messageRepository
}

func (suite *ListMessagesInteractorTestSuite) TearDownTest() {
	suite.messageRepository.AssertExpectations(suite.T())
}

func (suite *ListMessagesInteractorTestSuite) getValidRequest() request.ListMessages {
	return request.ListMessages{
		User: entity.User{
			Email: "test@test.com",
		},
	}
}

func (suite *ListMessagesInteractorTestSuite) TestRepositoryAnyError() {
	request := suite.getValidRequest()

	suite.messageRepository.On("AllWithUser", request.User.Email).Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *ListMessagesInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()
	messages := []entity.Message{
		{From: "a@b.com"},
		{From: "b@b.com"},
	}

	suite.messageRepository.On("AllWithUser", request.User.Email).Return(messages, nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.ListMessages{
		Total: 2,
		Items: []response.Message{
			{From: "a@b.com"},
			{From: "b@b.com"},
		},
	}
	suite.Equal(expected, r)
}
