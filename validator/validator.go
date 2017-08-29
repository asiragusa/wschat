// The package validator is used to validate the requests
package validator

import (
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"strings"
)

// Interface used mainly for Unit testing
type RequestValidator interface {
	Struct(interface{}) error
	FormatError(interface{}) response.Response
}

// The Request validator configures a gopkg.in/go-playground/validator.v9 validator to validate the requests
type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	// Used to extract the field name by the json tag
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag == "" || jsonTag == "-" {
			return field.Name
		}
		return jsonTag
	})

	return &Validator{
		validator: v,
	}
}

// Validates a given request
func (v Validator) Struct(i interface{}) error {
	return v.validator.Struct(i)
}

// Creates a Unprocessable Entity Response, given a validation error
func (v Validator) FormatError(err interface{}) response.Response {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return response.NewError(iris.StatusInternalServerError)
	}
	error := response.NewError(iris.StatusUnprocessableEntity)
	for _, err := range err.(validator.ValidationErrors) {
		error.AddDetail(err.Field(), err.Tag())
	}

	return error
}
