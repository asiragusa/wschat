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

type RegisterControllerTestSuite struct {
	suite.Suite
	controller *Register
	interactor *mocks.RegisterInteractor
	validator  *mocks.RequestValidator
	e          *httpexpect.Expect
}

func TestRegisterController(t *testing.T) {
	suite.Run(t, new(RegisterControllerTestSuite))
}

func (suite *RegisterControllerTestSuite) SetupSuite() {
	suite.controller = NewRegisterController()

	app := iris.New()
	app.Post("/", suite.controller.Handle)
	suite.e = httptest.New(suite.T(), app)
}

func (suite *RegisterControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.RegisterInteractor{}
	suite.validator = &mocks.RequestValidator{}

	suite.controller.Interactor = suite.interactor
	suite.controller.Validator = suite.validator
}

func (suite *RegisterControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
	suite.validator.AssertExpectations(suite.T())
}

func (suite *RegisterControllerTestSuite) validJSON() map[string]interface{} {
	return map[string]interface{}{
		"email":    "test@test.com",
		"password": "validPassword",
	}
}

func (suite *RegisterControllerTestSuite) requestObject() request.Request {
	return request.Register{
		Email:    "test@test.com",
		Password: "validPassword",
	}
}

func (suite *RegisterControllerTestSuite) validResponse() response.Response {
	return response.Register{
		AccessToken: "accessToken",
	}
}

func (suite *RegisterControllerTestSuite) TestBadRequest() {
	suite.e.POST("/").WithText("bad request").Expect().Status(httptest.StatusBadRequest)
}

func (suite *RegisterControllerTestSuite) TestUnprocessableEntity() {
	request := suite.requestObject()
	err := validator.ValidationErrors{}
	suite.validator.On("Struct", request).Return(err)
	suite.validator.On("FormatError", err).Return(response.NewError(httptest.StatusUnprocessableEntity))
	suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(httptest.StatusUnprocessableEntity)
}

func (suite *RegisterControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.validator.On("Struct", request).Return(nil)
	suite.interactor.On("Call", request).Return(response)

	r := suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(response.GetCode())
	r.JSON().Equal(response)
}
