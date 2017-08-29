// The package request defines all the requests accepted by the HTTP and WS endpoints
package request

import "github.com/asiragusa/wschat/entity"

type (
	// Base interface
	Request interface{}

	// Used by the POST /register endpoint
	Register struct {
		// User email
		Email string `json:"email" validate:"required,email"`

		// User password
		Password string `json:"password" validate:"required,min=6"`
	}

	// Used by POST /login
	Login struct {
		// User email
		Email string `json:"email" validate:"required"`

		// User password
		Password string `json:"password" validate:"required"`
	}

	// Used by POST /message and WS
	CreateMessage struct {
		// This field is assigned by the request handler. It represents the current authorized user
		From entity.User `json:"-"`

		// Message to
		To string `json:"to" validate:"required"`

		// Message text
		Message string `json:"message" validate:"required"`
	}

	// Used by GET /messages
	ListMessages struct {
		// This field is assigned by the request handler. It represents the current authorized user
		User entity.User
	}

	// Used by GET /users
	ListUsers struct {
	}

	// Used by POST /wsToken
	CreateWsToken struct {
		// This field is assigned by the request handler. It represents the current authorized user
		User entity.User
	}
)
