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
type LoginInteractor interface {
	Call(request.Login) response.Response
}

// Logs in an user. Returns the access token for authenticating the following requests
type Login struct {
	// Injected via DI
	UserRepository repository.UserRepository `inject:""`

	// Injected via DI
	AccessTokenGenerator services.TokenGenerator `inject:"accessTokenGenerator"`
}

func NewLoginInteractor() *Login {
	return &Login{}
}

func (i Login) Call(request request.Login) response.Response {
	// Find the user in the DB
	user, err := i.UserRepository.Login(request.Email, request.Password)
	if err == repository.UserBadUsernameOrPasswordError {
		return response.NewError(iris.StatusUnauthorized)
	}

	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	// Generate the access token
	token, err := i.AccessTokenGenerator.GenerateToken(*user, time.Hour*24*30)
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	return response.Login{
		AccessToken: token,
	}
}
