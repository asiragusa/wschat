package controller

import (
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/validator"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// Request handler for POST /messages
type CreateMessage struct {
	// Injected via DI
	Validator validator.RequestValidator `inject:""`

	// Injected via DI
	Interactor interactor.CreateMessageInteractor `inject:""`
}

func NewCreateMessageController() *CreateMessage {
	return &CreateMessage{}
}

func (c *CreateMessage) Handle(ctx context.Context) {
	request := request.CreateMessage{}

	if err := ctx.ReadJSON(&request); err != nil {
		sendResponse(ctx, response.NewError(iris.StatusBadRequest))
		return
	}

	request.From = *(ctx.Values().Get("user").(*entity.User))

	if err := c.Validator.Struct(request); err != nil {
		sendResponse(ctx, c.Validator.FormatError(err))
		return
	}

	sendResponse(ctx, c.Interactor.Call(request))
}
