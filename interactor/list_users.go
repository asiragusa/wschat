package interactor

import (
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris"
)

// Interface used mainly for Unit testing
type ListUsersInteractor interface {
	Call(request.ListUsers) response.Response
}

// Lists all the registered users
type ListUsers struct {
	// Injected via DI
	UserRepository repository.UserRepository `inject:""`
}

func NewListUsersInteractor() *ListUsers {
	return &ListUsers{}
}

func (i ListUsers) Call(request request.ListUsers) response.Response {
	// Find all the users
	users, err := i.UserRepository.All()
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	res := response.ListUsers{
		Total: len(users),
		Items: []response.User{},
	}

	for _, user := range users {
		res.Items = append(res.Items, response.User{
			Email: user.Email,
		})
	}
	return res
}
