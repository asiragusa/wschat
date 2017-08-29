package controller

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/kataras/iris/context"
)

// Request handler for GET /messages
type ListMessages struct {
	// Injected via DI
	Interactor interactor.ListMessagesInteractor `inject:""`
}

func NewListMessagesController() *ListMessages {
	return &ListMessages{}
}

func (c *ListMessages) Handle(ctx context.Context) {
	request := request.ListMessages{}
	request.User = *(ctx.Values().Get("user").(*entity.User))

	sendResponse(ctx, c.Interactor.Call(request))
}
