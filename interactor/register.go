package interactor

import (
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/kataras/iris"
	"time"
)

// Interface used mainly for Unit testing
type RegisterInteractor interface {
	Call(request.Register) response.Response
}

// Registers a new user. Returns the Access Token
type Register struct {
	// Injected via DI
	UserRepository repository.UserRepository `inject:""`

	// Injected via DI
	AccessTokenGenerator services.TokenGenerator `inject:"accessTokenGenerator"`
}

func NewRegisterInteractor() *Register {
	return &Register{}
}

func (i Register) Call(request request.Register) response.Response {
	// Create a new user in the DB
	user, err := i.UserRepository.CreateUser(request.Email, request.Password)
	if err == repository.UserAlreadyExistsError {
		error := response.NewError(iris.StatusUnprocessableEntity)
		error.AddDetail("email", "alreadyExists")
		return error
	}

	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	// Generate the Access Token
	token, err := i.AccessTokenGenerator.GenerateToken(*user, time.Hour*24*30)
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	return response.Register{
		AccessToken: token,
	}
}
