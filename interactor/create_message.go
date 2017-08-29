// The package interactor contains all the interactors used by request handlers
//
// This allows to separate business logic from HTTP handling. Interactors are also be used by websockets
package interactor

import (
	"fmt"
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/kataras/iris"
)

// Interface used mainly for Unit testing
type CreateMessageInteractor interface {
	Call(message request.CreateMessage) response.Response
}

// Interactor used to create a new message
type CreateMessage struct {
	// Injected via DI
	UserRepository repository.UserRepository `inject:""`

	// Injected via DI
	MessageRepository repository.MessageRepository `inject:""`

	// Injected via DI
	PubsubClient services.PubsubClient `inject:""`
}

func NewCreateMessageInteractor() *CreateMessage {
	return &CreateMessage{}
}

func (i CreateMessage) Call(request request.CreateMessage) response.Response {
	// Find the destination user or return an error
	to, err := i.UserRepository.GetUserByEmail(request.To)
	if err == repository.UserNotFoundError {
		error := response.NewError(iris.StatusUnprocessableEntity)
		error.AddDetail("to", "notExists")
		return error
	}
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	// Create the new message
	message, err := i.MessageRepository.Create(request.From.Email, to.Email, request.Message)
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	// Dispatch the message to the pubsub clients
	if err := i.PubsubClient.Publish(*message); err != nil {
		// TODO: do proper logging
		fmt.Println(err.Error())
	}

	return response.CreateMessage{
		Id:        message.Id,
		From:      request.From.Email,
		To:        to.Email,
		Message:   message.Message,
		CreatedAt: message.CreatedAt,
	}
}
