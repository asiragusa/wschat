package interactor

import (
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/kataras/iris"
	"time"
)

// Interface used mainly for Unit testing
type WsTokenInteractor interface {
	Call(request.CreateWsToken) response.Response
}

// Creates a new token to authentificate the user via websocket
type WsToken struct {
	// Injected via DI
	WsTokenGenerator services.TokenGenerator `inject:"wsTokenGenerator"`
}

func NewWsTokenInteractor() *WsToken {
	return &WsToken{}
}

func (i WsToken) Call(request request.CreateWsToken) response.Response {
	// Generates a new token for the given (already validated) user
	token, err := i.WsTokenGenerator.GenerateToken(request.User, time.Second*30)
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	return response.CreateWsToken{
		Token: token,
	}
}
