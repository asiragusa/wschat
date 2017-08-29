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

type ListMessagesControllerTestSuite struct {
	suite.Suite
	controller *ListMessages
	interactor *mocks.ListMessagesInteractor
	user       *entity.User
	e          *httpexpect.Expect
}

func TestListMessagesController(t *testing.T) {
	suite.Run(t, new(ListMessagesControllerTestSuite))
}

func (suite *ListMessagesControllerTestSuite) SetupSuite() {
	suite.controller = NewListMessagesController()
	suite.user = &entity.User{
		Email: "a@b.com",
	}

	app := iris.New()
	app.Use(func(ctx context.Context) {
		ctx.Values().Set("user", suite.user)
		ctx.Next()
	})
	app.Get("/", suite.controller.Handle)
	suite.e = httptest.New(suite.T(), app)
}

func (suite *ListMessagesControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.ListMessagesInteractor{}

	suite.controller.Interactor = suite.interactor
}

func (suite *ListMessagesControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
}

func (suite *ListMessagesControllerTestSuite) requestObject() request.Request {
	return request.ListMessages{
		User: *suite.user,
	}
}

func (suite *ListMessagesControllerTestSuite) validResponse() response.Response {
	return response.ListMessages{
		Total: 1,
		Items: []response.Message{{
			From: "a@b.com",
			To:   "b@b.com",
		}},
	}
}

func (suite *ListMessagesControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.interactor.On("Call", request).Return(response)

	r := suite.e.GET("/").Expect().Status(response.GetCode())
	r.JSON().Equal(response)
}
