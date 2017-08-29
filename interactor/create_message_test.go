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

type CreateMessageInteractorTestSuite struct {
	suite.Suite
	interactor        *CreateMessage
	userRepository    *mocks.UserRepository
	messageRepository *mocks.MessageRepository
	pubsubClient      *mocks.PubsubClient
}

func TestCreateMessageInteractor(t *testing.T) {
	suite.Run(t, new(CreateMessageInteractorTestSuite))
}

func (suite *CreateMessageInteractorTestSuite) SetupSuite() {
	suite.interactor = NewCreateMessageInteractor()
}

func (suite *CreateMessageInteractorTestSuite) SetupTest() {
	suite.userRepository = &mocks.UserRepository{}
	suite.messageRepository = &mocks.MessageRepository{}
	suite.pubsubClient = &mocks.PubsubClient{}

	suite.interactor.UserRepository = suite.userRepository
	suite.interactor.MessageRepository = suite.messageRepository
	suite.interactor.PubsubClient = suite.pubsubClient
}

func (suite *CreateMessageInteractorTestSuite) TearDownTest() {
	suite.userRepository.AssertExpectations(suite.T())
	suite.messageRepository.AssertExpectations(suite.T())
	suite.pubsubClient.AssertExpectations(suite.T())
}

func (suite *CreateMessageInteractorTestSuite) getValidRequest() request.CreateMessage {
	return request.CreateMessage{
		From: entity.User{
			Email: "a@b.com",
		},
		To:      "b@b.com",
		Message: "test",
	}
}

func (suite *CreateMessageInteractorTestSuite) TestToNotFound() {
	request := suite.getValidRequest()

	suite.userRepository.On("GetUserByEmail", request.To).Return(nil, repository.UserNotFoundError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.NewError(httptest.StatusUnprocessableEntity)
	expected.AddDetail("to", "notExists")

	suite.Equal(expected, r)
}

func (suite *CreateMessageInteractorTestSuite) TestGetToAnError() {
	request := suite.getValidRequest()

	suite.userRepository.On("GetUserByEmail", request.To).Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *CreateMessageInteractorTestSuite) TestMessageRepositoryAnError() {
	request := suite.getValidRequest()

	suite.userRepository.On("GetUserByEmail", request.To).Return(&entity.User{
		Email: request.To,
	}, nil)
	suite.messageRepository.On("Create", request.From.Email, request.To, request.Message).Return(nil, assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *CreateMessageInteractorTestSuite) TestPubsubAnyError() {
	request := suite.getValidRequest()
	message := entity.Message{
		Id:        "messageId",
		From:      request.From.Email,
		To:        request.To,
		Message:   request.Message,
		CreatedAt: time.Now(),
	}

	suite.userRepository.On("GetUserByEmail", request.To).Return(&entity.User{
		Email: request.To,
	}, nil)
	suite.messageRepository.On("Create", request.From.Email, request.To, request.Message).Return(&message, nil)
	suite.pubsubClient.On("Publish", message).Return(assert.AnError)

	// Here we should test that we logged the error

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.CreateMessage{
		Id:        message.Id,
		From:      message.From,
		To:        message.To,
		Message:   message.Message,
		CreatedAt: message.CreatedAt,
	}
	suite.Equal(expected, r)
}

func (suite *CreateMessageInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()
	message := entity.Message{
		Id:        "messageId",
		From:      request.From.Email,
		To:        request.To,
		Message:   request.Message,
		CreatedAt: time.Now(),
	}

	suite.userRepository.On("GetUserByEmail", request.To).Return(&entity.User{
		Email: request.To,
	}, nil)
	suite.messageRepository.On("Create", request.From.Email, request.To, request.Message).Return(&message, nil)
	suite.pubsubClient.On("Publish", message).Return(nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.CreateMessage{
		Id:        message.Id,
		From:      message.From,
		To:        message.To,
		Message:   message.Message,
		CreatedAt: message.CreatedAt,
	}
	suite.Equal(expected, r)
}
