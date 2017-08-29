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
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

type CreateMessageControllerTestSuite struct {
	suite.Suite
	controller *CreateMessage
	interactor *mocks.CreateMessageInteractor
	validator  *mocks.RequestValidator
	user       *entity.User
	e          *httpexpect.Expect
}

func TestCreateMessageController(t *testing.T) {
	suite.Run(t, new(CreateMessageControllerTestSuite))
}

func (suite *CreateMessageControllerTestSuite) SetupSuite() {
	suite.controller = NewCreateMessageController()
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

func (suite *CreateMessageControllerTestSuite) SetupTest() {
	suite.interactor = &mocks.CreateMessageInteractor{}
	suite.validator = &mocks.RequestValidator{}

	suite.controller.Interactor = suite.interactor
	suite.controller.Validator = suite.validator
}

func (suite *CreateMessageControllerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
	suite.validator.AssertExpectations(suite.T())
}

func (suite *CreateMessageControllerTestSuite) validJSON() map[string]interface{} {
	return map[string]interface{}{
		"to":      "test@test.com",
		"message": "test",
	}
}

func (suite *CreateMessageControllerTestSuite) requestObject() request.Request {
	return request.CreateMessage{
		From:    *suite.user,
		To:      "test@test.com",
		Message: "test",
	}
}

func (suite *CreateMessageControllerTestSuite) validResponse() response.Response {
	return response.NoContentResponse{}
}

func (suite *CreateMessageControllerTestSuite) TestBadRequest() {
	suite.e.POST("/").WithText("bad request").Expect().Status(httptest.StatusBadRequest)
}

func (suite *CreateMessageControllerTestSuite) TestUnprocessableEntity() {
	request := suite.requestObject()
	err := validator.ValidationErrors{}
	suite.validator.On("Struct", request).Return(err)
	suite.validator.On("FormatError", err).Return(response.NewError(httptest.StatusUnprocessableEntity))
	suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(httptest.StatusUnprocessableEntity)
}

func (suite *CreateMessageControllerTestSuite) TestHandleOk() {
	request := suite.requestObject()
	response := suite.validResponse()

	suite.validator.On("Struct", request).Return(nil)
	suite.interactor.On("Call", request).Return(response)

	r := suite.e.POST("/").WithJSON(suite.validJSON()).Expect().Status(response.GetCode())
	r.Body().Empty()
}
