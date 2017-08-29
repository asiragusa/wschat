package controller

import (
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/iris-contrib/httpexpect"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

type LoginControllerTestSuite struct {
	suite.Suite
	controller *Login
	interactor *mocks.LoginInteractor
	validator  *mocks.RequestValidator
	e          *httpexpect.Expect
}

func TestLoginController(t *testing.T) {
	suite.Run(t, new(LoginControllerTestSuite))
}

func (suite *LoginControllerTestSuite) SetupSuite() {
	suite.controller = NewLoginController()

	app := iris.New()
	app.Post("/", suite.controller.Handle)
	suite.e = httptest.New(suite.T(), app)
}

func (suite *LoginControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.LoginInteractor{}
	suite.validator = &mocks.RequestValidator{}

	suite.controller.Interactor = suite.interactor
	suite.controller.Validator = suite.validator
}

func (suite *LoginControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
	suite.validator.AssertExpectations(suite.T())
}

func (suite *LoginControllerTestSuite) validJSON() map[string]interface{} {
	return map[string]interface{}{
		"email":    "test@test.com",
		"password": "validPassword",
	}
}

func (suite *LoginControllerTestSuite) requestObject() request.Request {
	return request.Login{
		Email:    "test@test.com",
		Password: "validPassword",
	}
}

func (suite *LoginControllerTestSuite) validResponse() response.Response {
	return response.Login{
		AccessToken: "accessToken",
	}
}

func (suite *LoginControllerTestSuite) TestBadRequest() {
	suite.e.POST("/").WithText("bad request").Expect().Status(httptest.StatusBadRequest)
}

func (suite *LoginControllerTestSuite) TestUnprocessableEntity() {
	request := suite.requestObject()
	err := validator.ValidationErrors{}
	suite.validator.On("Struct", request).Return(err)
	suite.validator.On("FormatError", err).Return(response.NewError(httptest.StatusUnprocessableEntity))
	suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(httptest.StatusUnprocessableEntity)
}

func (suite *LoginControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.validator.On("Struct", request).Return(nil)
	suite.interactor.On("Call", request).Return(response)

	r := suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(response.GetCode())
	r.JSON().Equal(response)
}
