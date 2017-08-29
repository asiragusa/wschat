package controller

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/kataras/iris/context"
)

// Request handler for POST /wsToken
type WsToken struct {
	// Injected via DI
	Interactor interactor.WsTokenInteractor `inject:""`
}

func NewWsTokenController() *WsToken {
	return &WsToken{}
}

func (c *WsToken) Handle(ctx context.Context) {
	request := request.CreateWsToken{}
	request.User = *(ctx.Values().Get("user").(*entity.User))

	sendResponse(ctx, c.Interactor.Call(request))
}
