package controller

import (
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/validator"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// Request handler for POST /login
type Login struct {
	// Injected via DI
	Validator validator.RequestValidator `inject:""`

	// Injected via DI
	Interactor interactor.LoginInteractor `inject:""`
}

func NewLoginController() *Login {
	return &Login{}
}

func (c *Login) Handle(ctx context.Context) {
	request := request.Login{}
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
