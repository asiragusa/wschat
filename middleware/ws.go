package middleware

import (
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// Authenticate a websocket connection via JWT token sent via the query param token
//
// Eg. ws(s)://localhost/ws?token=TOKEN
type Ws struct {
	WsTokenGenerator services.TokenGenerator `inject:"wsTokenGenerator"`
}

func NewWsMiddleware() *Ws {
	return &Ws{}
}

func (m *Ws) Handle(ctx context.Context) {
	// Get the param
	token := ctx.URLParam("token")

	// Validate the token
	user, err := m.WsTokenGenerator.ValidateToken(token)

	// Invalid token
	if err == services.InvalidTokenError {
		error := response.NewError(iris.StatusUnauthorized)
		ctx.StatusCode(error.GetCode())
		ctx.JSON(error)
		ctx.StopExecution()
		return
	}

	// Unexpected error
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
