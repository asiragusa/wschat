// The package response defines all the responses returned by the interactors
package response

import (
	"github.com/kataras/iris"
	"time"
)

type (
	// Response interface
	Response interface {
		// Returns the HTTP code
		GetCode() int
	}

	// 200 Response
	OKResponse struct{}
	// 201 Response
	CreatedResponse struct{}
	// 204 Response
	NoContentResponse struct{}

	// Used by POST /register endpoint
	Register struct {
		// Returns 204
		CreatedResponse
		// Access token
		AccessToken string `json:"accessToken"`
	}

	// Used by POST /login endpoint
	Login struct {
		// Returns 200
		OKResponse

		// Access token
		AccessToken string `json:"accessToken"`
	}

	// Contains a message
	Message struct {
		Id        string    `json:"id"`
		From      string    `json:"from"`
		To        string    `json:"to"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"createdAt"`
	}

	// Used by GET /messages endpoint
	ListMessages struct {
		// Returns 200
		OKResponse

		// Total items
		Total int `json:"total"`

		// Array of found messages
		Items []Message `json:"items"`
	}

	// Used by POST /messages endpoint and WS
	CreateMessage struct {
		// Returns 201
		CreatedResponse

		// Message ID
		Id string `json:"id"`

		// Message from email
		From string `json:"from"`

		// Message to email
		To string `json:"to"`

		// Message text
		Message string `json:"message"`

		// Created At
		CreatedAt time.Time `json:"createdAt"`
	}

	// Used by GET /users endpoint
	User struct {
		// User email
		Email string `json:"email"`
	}

	// Used by GET /users endpoint
	ListUsers struct {
		// Returns 200
		OKResponse
		// Total items
		Total int `json:"total"`

		// User list
		Items []User `json:"items"`
	}

	// Used by POST /wsToken
	CreateWsToken struct {
		// Returns 201
		CreatedResponse

		// WS token
		Token string `json:"token"`
	}
)

// Return 200
func (r OKResponse) GetCode() int {
	return iris.StatusOK
}

// Return 201
func (r CreatedResponse) GetCode() int {
	return iris.StatusCreated
}

// Return 204
func (r NoContentResponse) GetCode() int {
	return iris.StatusNoContent
}
