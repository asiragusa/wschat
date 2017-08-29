package controller

import (
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/validator"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// Request handler for POST /register
type Register struct {
	// Injected via DI
	Validator validator.RequestValidator `inject:""`

	// Injected via DI
	Interactor interactor.RegisterInteractor `inject:""`
}

func NewRegisterController() *Register {
	return &Register{}
}

func (c *Register) Handle(ctx context.Context) {
	request := request.Register{}
	if err := ctx.ReadJSON(&request); err != nil {
		sendResponse(ctx, response.NewError(iris.StatusBadRequest))
		return
	}

	if err := c.Validator.Struct(request); err != nil {
		sendResponse(ctx, c.Validator.FormatError(err))
		return
	}

	sendResponse(ctx, c.Interactor.Call(request))
}
