package controller

import (
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/kataras/iris/context"
)

// Request handler for GET /users
type ListUsers struct {
	// Injected via DI
	Interactor interactor.ListUsersInteractor `inject:""`
}

func NewListUsersController() *ListUsers {
	return &ListUsers{}
}

func (c *ListUsers) Handle(ctx context.Context) {
	request := request.ListUsers{}

	sendResponse(ctx, c.Interactor.Call(request))
}
