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

type AuthenticatedMiddlewareTestSuite struct {
	suite.Suite
	middleware *Authenticated
	generator  *mocks.TokenGenerator
	e          *httpexpect.Expect
}

func TestAuthenticatedMiddleware(t *testing.T) {
	suite.Run(t, new(AuthenticatedMiddlewareTestSuite))
}

func (suite *AuthenticatedMiddlewareTestSuite) SetupSuite() {
	suite.middleware = NewAuthenticatedMiddleware()

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

func (suite *AuthenticatedMiddlewareTestSuite) SetupTest() {
	suite.generator = &mocks.TokenGenerator{}

	suite.middleware.AccessTokenGenerator = suite.generator
}

func (suite *AuthenticatedMiddlewareTestSuite) TearDownTest() {
	suite.generator.AssertExpectations(suite.T())
}

func (suite *AuthenticatedMiddlewareTestSuite) TestBadHeader() {
	headers := []string{
		"",
		"A",
		"Basic",
		"Bearer ",
		"Basic A",
		"Bearer A B",
	}

	suite.generator.On("ValidateToken", "").Times(len(headers)).Return(nil, services.InvalidTokenError)

	for _, auth := range headers {
		suite.e.GET("/").WithHeader("Authorization", auth).Expect().Status(httptest.StatusUnauthorized).
			JSON().Object().Equal(map[string]interface{}{
			"code":    httptest.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}
}

func (suite *AuthenticatedMiddlewareTestSuite) TestBadAuth() {
	auth := "Bearer invalid"

	suite.generator.On("ValidateToken", "invalid").Return(nil, services.InvalidTokenError)
	suite.e.GET("/").WithHeader("Authorization", auth).Expect().Status(httptest.StatusUnauthorized).
		JSON().Object().Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

func (suite *AuthenticatedMiddlewareTestSuite) TestGeneratorAnError() {
	auth := "Bearer valid"

	suite.generator.On("ValidateToken", "valid").Return(nil, assert.AnError)
	suite.e.GET("/").WithHeader("Authorization", auth).Expect().Status(httptest.StatusInternalServerError).
		JSON().Object().Equal(map[string]interface{}{
		"code":    httptest.StatusInternalServerError,
		"message": "Internal Server Error",
	})
}

func (suite *AuthenticatedMiddlewareTestSuite) TestHandleOk() {
	auth := "Bearer valid"

	suite.generator.On("ValidateToken", "valid").Return(&entity.User{
		Email: "test",
	}, nil)
	suite.e.GET("/").WithHeader("Authorization", auth).Expect().Status(httptest.StatusOK)
}
