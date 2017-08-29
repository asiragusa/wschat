package interactor

import (
	"github.com/asiragusa/wschat/repository"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris"
)

// Interface used mainly for Unit testing
type ListMessagesInteractor interface {
	Call(request.ListMessages) response.Response
}

// ListMessages returns all the messages for the given user
type ListMessages struct {
	// Injected via DI
	MessageRepository repository.MessageRepository `inject:""`
}

func NewListMessagesInteractor() *ListMessages {
	return &ListMessages{}
}

func (i ListMessages) Call(request request.ListMessages) response.Response {
	// Find messages
	messages, err := i.MessageRepository.AllWithUser(request.User.Email)
	if err != nil {
		return response.NewError(iris.StatusInternalServerError)
	}

	res := response.ListMessages{
		Total: len(messages),
		Items: []response.Message{},
	}

	for _, message := range messages {
		res.Items = append(res.Items, response.Message{
			Id:        message.Id,
			From:      message.From,
			To:        message.To,
			Message:   message.Message,
			CreatedAt: message.CreatedAt,
		})
	}
	return res
}
