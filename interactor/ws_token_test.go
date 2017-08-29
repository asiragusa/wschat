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
	"time"
)

type WsTokenInteractorTestSuite struct {
	suite.Suite
	interactor *WsToken
	generator  *mocks.TokenGenerator
}

func TestWsTokenInteractor(t *testing.T) {
	suite.Run(t, new(WsTokenInteractorTestSuite))
}

func (suite *WsTokenInteractorTestSuite) SetupSuite() {
	suite.interactor = NewWsTokenInteractor()
}

func (suite *WsTokenInteractorTestSuite) SetupTest() {
	suite.generator = &mocks.TokenGenerator{}

	suite.interactor.WsTokenGenerator = suite.generator
}

func (suite *WsTokenInteractorTestSuite) TearDownTest() {
	suite.generator.AssertExpectations(suite.T())
}

func (suite *WsTokenInteractorTestSuite) getValidRequest() request.CreateWsToken {
	return request.CreateWsToken{
		User: entity.User{
			Email: "test@test.com",
		},
	}
}

func (suite *WsTokenInteractorTestSuite) TestGeneratorAnyError() {
	request := suite.getValidRequest()

	suite.generator.On("GenerateToken", request.User, time.Second*30).Return("", assert.AnError)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)
	suite.Equal(response.NewError(httptest.StatusInternalServerError), r)
}

func (suite *WsTokenInteractorTestSuite) TestOK() {
	request := suite.getValidRequest()

	suite.generator.On("GenerateToken", request.User, time.Second*30).Return("token", nil)

	r := suite.interactor.Call(request)
	suite.Require().NotNil(r)

	expected := response.CreateWsToken{
		Token: "token",
	}
	suite.Equal(expected, r)
}
