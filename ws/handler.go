// This package contains the handler for the websocket connections
package ws

import (
	"encoding/json"
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/interactor"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/asiragusa/wschat/services"
	"github.com/asiragusa/wschat/validator"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

// Websocket response, used to wrap the response.Response.
// If a requestId is given in the request it is returned to identify the response to the corresponding request
type WsResponse struct {
	// RequestId as in the request
	RequestId string `json:"requestId,omitempty"`

	// Response content
	Body interface{} `json:"body"`
}

// Handler for the websocket requests
type Handler struct {
	// Injected via DI
	PubsubClient services.PubsubClient `inject:""`

	// Injected via DI
	CreateMessageInteractor interactor.CreateMessageInteractor `inject:""`

	// Injected via DI
	Validator validator.RequestValidator `inject:""`
}

func NewWsHandler() *Handler {
	return &Handler{}
}

// Parse the received message
// Returns the requestId and a bool true if the message has been correctly parsed
func (h *Handler) parseRequest(c websocket.Connection, msg interface{}, dst interface{}) (string, bool) {
	// Ignore the message is not a map[string]interface{}
	m, ok := msg.(map[string]interface{})
	if !ok {
		return "", false
	}

	// Return an error if the requestId can't be casted to string
	requestId, ok := m["requestId"].(string)
	if !ok {
		return "", false
	}

	// Coerce msg into req
	encoded, err := json.Marshal(m["body"])
	if err != nil {
		c.Emit("error", WsResponse{
			RequestId: requestId,
			Body:      response.NewError(iris.StatusBadRequest),
		})
		return "", false
	}

	if err := json.Unmarshal(encoded, dst); err != nil {
		c.Emit("error", WsResponse{
			RequestId: requestId,
			Body:      response.NewError(iris.StatusBadRequest),
		})
		return "", false
	}
	return requestId, true
}

// Handle the `message` request
func (h *Handler) handleMessage(c websocket.Connection, requestId string, req request.CreateMessage) {
	// Validate the request
	if err := h.Validator.Struct(req); err != nil {
		c.Emit("error", WsResponse{
			RequestId: requestId,
			Body:      h.Validator.FormatError(err),
		})
		return
	}

	// Create the message
	res := h.CreateMessageInteractor.Call(req)

	var successResponse response.CreateMessage
	// If the response code is not the correct one return the error
	if res.GetCode() != successResponse.GetCode() {
		c.Emit("error", WsResponse{
			RequestId: requestId,
			Body:      res,
		})
	}

	// Send a message to confirm that the message has been sent
	c.Emit("sent", WsResponse{
		RequestId: requestId,
		Body:      res,
	})
}

// Websocket connection handler
func (h *Handler) HandleConnection(c websocket.Connection) {
	// Fetch the user from the request
	user := c.Context().Values().Get("user").(*entity.User)

	// Subscribe to the messages for user.Email
	err, cancelFn := h.PubsubClient.Subscribe(user.Email, func(message entity.Message) {
		c.Emit("message", WsResponse{
			Body: response.CreateMessage{
				Id:        message.Id,
				From:      message.From,
				To:        message.To,
				Message:   message.Message,
				CreatedAt: message.CreatedAt,
			},
		})
	})

	if err != nil {
		// TODO: Find out why disconnect panics
		//c.Disconnect()
		return
	}

	// Handler for the message request
	c.On("message", func(msg interface{}) {
		var req request.CreateMessage

		// Parse the request
		requestId, ok := h.parseRequest(c, msg, &req)
		if !ok {
			return
		}

		//
		req.From = *user

		h.handleMessage(c, requestId, req)
	})

	// Handler for the disconnection. Calls the cancelFn of the subscription
	c.OnDisconnect(func() {
		cancelFn()
	})
}
