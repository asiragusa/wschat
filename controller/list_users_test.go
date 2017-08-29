package controller

import (
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/iris-contrib/httpexpect"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ListUsersControllerTestSuite struct {
	suite.Suite
	controller *ListUsers
	interactor *mocks.ListUsersInteractor
	e          *httpexpect.Expect
}

func TestListUsersController(t *testing.T) {
	suite.Run(t, new(ListUsersControllerTestSuite))
}

func (suite *ListUsersControllerTestSuite) SetupSuite() {
	suite.controller = NewListUsersController()

	app := iris.New()
	app.Get("/", suite.controller.Handle)
	suite.e = httptest.New(suite.T(), app)
}

func (suite *ListUsersControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.ListUsersInteractor{}

	suite.controller.Interactor = suite.interactor
}

func (suite *ListUsersControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
}

func (suite *ListUsersControllerTestSuite) requestObject() request.Request {
	return request.ListUsers{}
}

func (suite *ListUsersControllerTestSuite) validResponse() response.Response {
	return response.ListUsers{
		Total: 1,
		Items: []response.User{{
			Email: "a@b.com",
		}},
	}
}

func (suite *ListUsersControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.interactor.On("Call", request).Return(response)

	r := suite.e.GET("/").Expect().Status(response.GetCode())
	r.JSON().Equal(response)
}
