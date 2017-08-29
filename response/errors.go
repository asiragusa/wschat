package response

import "github.com/kataras/iris"

var messages = map[int]string{
	iris.StatusBadRequest:          "Bad Request",
	iris.StatusUnauthorized:        "Unauthorized",
	iris.StatusForbidden:           "Forbidden",
	iris.StatusNotFound:            "Not Found",
	iris.StatusUnprocessableEntity: "Unprocessable Entity",
	iris.StatusInternalServerError: "Internal Server Error",
}

// Error response type
type Error struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}

// Returns the error code
func (e *Error) GetCode() int {
	return e.Code
}

// Adds a detail to the error
func (e *Error) AddDetail(field, detail string) {
	if e.Details[field] == nil {
		e.Details[field] = []string{}
	}
	e.Details[field] = append(e.Details[field], detail)
}

// Creates a new Error Response with code `code`
func NewError(code int) *Error {
	// If code is not found, message is an empty string, which is acceptable
	return &Error{code, messages[code], make(map[string][]string)}
}
