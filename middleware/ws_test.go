package middleware

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/services"
	"github.com/iris-contrib/httpexpect"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type WsMiddlewareTestSuite struct {
	suite.Suite
	middleware *Ws
	generator  *mocks.TokenGenerator
	e          *httpexpect.Expect
}

func TestWsMiddleware(t *testing.T) {
	suite.Run(t, new(WsMiddlewareTestSuite))
}

func (suite *WsMiddlewareTestSuite) SetupSuite() {
	suite.middleware = NewWsMiddleware()

	app := iris.New()
	app.Use(suite.middleware.Handle)
	app.Get("/", func(ctx context.Context) {
		user := ctx.Values().Get("user").(*entity.User)
		suite.Require().NotNil(user)
		suite.Equal("test", user.Email)

		ctx.StatusCode(httptest.StatusOK)
		ctx.JSON(map[string]bool{
			"ok": true,
		})
	})
	suite.e = httptest.New(suite.T(), app)
}

func (suite *WsMiddlewareTestSuite) SetupTest() {
	suite.generator = &mocks.TokenGenerator{}

	suite.middleware.WsTokenGenerator = suite.generator
}

func (suite *WsMiddlewareTestSuite) TearDownTest() {
	suite.generator.AssertExpectations(suite.T())
}

func (suite *WsMiddlewareTestSuite) TestNoParam() {
	suite.generator.On("ValidateToken", "").Return(nil, services.InvalidTokenError)

	suite.e.GET("/").Expect().Status(httptest.StatusUnauthorized).
		JSON().Object().Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

func (suite *WsMiddlewareTestSuite) TestBadAuth() {
	token := "invalid"

	suite.generator.On("ValidateToken", "invalid").Return(nil, services.InvalidTokenError)
	suite.e.GET("/").WithQuery("token", token).Expect().Status(httptest.StatusUnauthorized).
		JSON().Object().Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

func (suite *WsMiddlewareTestSuite) TestGeneratorAnError() {
	token := "valid"

	suite.generator.On("ValidateToken", "valid").Return(nil, assert.AnError)
	suite.e.GET("/").WithQuery("token", token).Expect().Status(httptest.StatusInternalServerError).
		JSON().Object().Equal(map[string]interface{}{
		"code":    httptest.StatusInternalServerError,
		"message": "Internal Server Error",
	})
}

func (suite *WsMiddlewareTestSuite) TestOk() {
	token := "valid"

	suite.generator.On("ValidateToken", "valid").Return(&entity.User{
		Email: "test",
	}, nil)
	suite.e.GET("/").WithQuery("token", token).Expect().Status(httptest.StatusOK)
}
