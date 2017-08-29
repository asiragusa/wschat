package validator

import (
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

type ValidatorTestSuite struct {
	suite.Suite
	validator RequestValidator
}

func TestValidator(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

func (suite *ValidatorTestSuite) SetupSuite() {
	suite.validator = NewValidator()
}

type Test struct {
	F1 string `json:"f1" validate:"required"`
	F2 string `json:"f2,omitempty" validate:"required"`
	F3 string `json:"-" validate:"required"`
	F4 string `validate:"required"`
}

func (suite *ValidatorTestSuite) TestTagNameFunc() {
	err := suite.validator.Struct(&Test{})
	suite.Require().NotNil(err)

	errors := map[string]string{}
	for _, err := range err.(validator.ValidationErrors) {
		errors[err.StructField()] = err.Field()
	}

	expected := map[string]string{
		"F1": "f1",
		"F2": "f2",
		"F3": "F3",
		"F4": "F4",
	}

	suite.Equal(expected, errors)
}

func (suite *ValidatorTestSuite) TestFormatError() {
	err := suite.validator.Struct(&Test{})
	suite.Require().NotNil(err)

	res := suite.validator.FormatError(err)

	expected := response.NewError(httptest.StatusUnprocessableEntity)
	expected.AddDetail("f1", "required")
	expected.AddDetail("f2", "required")
	expected.AddDetail("F3", "required")
	expected.AddDetail("F4", "required")

	suite.Equal(expected, res)

	res2 := suite.validator.FormatError(&validator.InvalidValidationError{})
	suite.Equal(response.NewError(httptest.StatusInternalServerError), res2)
}
