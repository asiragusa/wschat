package request

import (
	"github.com/asiragusa/wschat/validator"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
)

type RequestsTestSuite struct {
	suite.Suite
	validator validator.RequestValidator
}

func TestRequests(t *testing.T) {
	suite.Run(t, new(RequestsTestSuite))
}

func (suite *RequestsTestSuite) SetupSuite() {
	suite.validator = validator.NewValidator()
}

func (suite *RequestsTestSuite) each(r interface{}, f func(interface{})) {
	requests := reflect.ValueOf(r)
	for i := 0; i < requests.Len(); i++ {
		f(requests.Index(i).Interface())
	}
}

func (suite *RequestsTestSuite) mustNotValidate(requests interface{}) {
	suite.each(requests, suite.mustNotValidateOne)
}

func (suite *RequestsTestSuite) mustValidate(requests interface{}) {
	suite.each(requests, suite.mustValidateOne)
}

func (suite *RequestsTestSuite) mustNotValidateOne(request interface{}) {
	err := suite.validator.Struct(request)
	suite.Error(err)
}

func (suite *RequestsTestSuite) mustValidateOne(request interface{}) {
	err := suite.validator.Struct(request)
	suite.NoError(err)
}

func (suite *RequestsTestSuite) TestRegisterInvalid() {
	suite.mustNotValidate([]*Register{
		{
		// Empty Request
		},
		{
			Email: "a@b.com",
		},
		{
			Password: "validPassword",
		},
		{
			Email:    "a",
			Password: "validPassword",
		},
		{
			Email:    "a@b.com",
			Password: "aaaaa",
		},
	})
}

func (suite *RequestsTestSuite) TestRegisterValid() {
	suite.mustValidateOne(Register{
		Email:    "a@b.com",
		Password: "aaaaaa",
	})
}

func (suite *RequestsTestSuite) TestLoginInvalid() {
	suite.mustNotValidate([]*Login{
		{
		// Empty Request
		},
		{
			Email: "a@b.com",
		},
		{
			Password: "validPassword",
		},
	})
}

func (suite *RequestsTestSuite) TestLoginValid() {
	suite.mustValidateOne(Login{
		Email:    "a@b.com",
		Password: "aaaaaa",
	})
}

func (suite *RequestsTestSuite) TestCreateMessageInvalid() {
	suite.mustNotValidate([]*CreateMessage{
		{
		// Empty Request
		},
		{
			To: "a@b.com",
		},
		{
			Message: "",
		},
	})
}

func (suite *RequestsTestSuite) TestCreateMessageValid() {
	suite.mustValidateOne(CreateMessage{
		To:      "a",
		Message: "a",
	})
}
