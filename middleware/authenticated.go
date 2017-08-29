// This package contains the middlewares used by the router
package middleware

import (
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"strings"
)

// Authenticates an user via JWT Token, sent with the Authorization header
type Authenticated struct {
	// Injected via DI
	AccessTokenGenerator services.TokenGenerator `inject:"accessTokenGenerator"`
}

func NewAuthenticatedMiddleware() *Authenticated {
	return &Authenticated{}
}

// Parses the header. Only Bearer authorization is allowed
func parseAuthorizationHeader(authorization string) string {
	if authorization == "" {
		return ""
	}
	splitted := strings.Split(authorization, " ")
	if len(splitted) != 2 {
		return ""
	}

	if splitted[0] != "Bearer" {
		return ""
	}

	return splitted[1]
}

// Middleware handler
func (m *Authenticated) Handle(ctx context.Context) {
	token := parseAuthorizationHeader(ctx.GetHeader("Authorization"))

	user, err := m.AccessTokenGenerator.ValidateToken(token)

	// If the token is not valid send 401
	if err == services.InvalidTokenError {
		error := response.NewError(iris.StatusUnauthorized)
		ctx.StatusCode(error.GetCode())
		ctx.JSON(error)
		ctx.StopExecution()
		return
	}

	// Internal Server Error
	if err != nil {
		error := response.NewError(iris.StatusInternalServerError)
		ctx.StatusCode(error.GetCode())
		ctx.JSON(error)
		ctx.StopExecution()
		return
	}

	// Set the user in the iris context
	ctx.Values().Set("user", user)
	ctx.Next()
}
