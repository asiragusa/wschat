package controller

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/iris-contrib/httpexpect"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/suite"
	"testing"
)

type WsTokenControllerTestSuite struct {
	suite.Suite
	controller *WsToken
	interactor *mocks.WsTokenInteractor
	user       *entity.User
	e          *httpexpect.Expect
}

func TestWsTokenController(t *testing.T) {
	suite.Run(t, new(WsTokenControllerTestSuite))
}

func (suite *WsTokenControllerTestSuite) SetupSuite() {
	suite.controller = NewWsTokenController()
	suite.user = &entity.User{
		Email: "a@b.com",
	}

	app := iris.New()
	app.Use(func(ctx context.Context) {
		ctx.Values().Set("user", suite.user)
		ctx.Next()
	})
	app.Post("/", suite.controller.Handle)
	suite.e = httptest.New(suite.T(), app)
}

func (suite *WsTokenControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.WsTokenInteractor{}

	suite.controller.Interactor = suite.interactor
}

func (suite *WsTokenControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
}

func (suite *WsTokenControllerTestSuite) requestObject() request.Request {
	return request.CreateWsToken{
		User: *suite.user,
	}
}

func (suite *WsTokenControllerTestSuite) validResponse() response.Response {
	return response.CreateWsToken{
		Token: "valid",
	}
}

func (suite *WsTokenControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.interactor.On("Call", request).Return(response)

	r := suite.e.POST("/").Expect().Status(response.GetCode())
	r.JSON().Equal(response)
}
