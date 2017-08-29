// Code generated by mockery v1.0.0
package mocks

import mock "github.com/stretchr/testify/mock"
import request "github.com/asiragusa/wschat/request"
import response "github.com/asiragusa/wschat/response"

// CreateMessageInteractor is an autogenerated mock type for the CreateMessageInteractor type
type CreateMessageInteractor struct {
	mock.Mock
}

// Call provides a mock function with given fields: message
func (_m *CreateMessageInteractor) Call(message request.CreateMessage) response.Response {
	ret := _m.Called(message)

	var r0 response.Response
	if rf, ok := ret.Get(0).(func(request.CreateMessage) response.Response); ok {
		r0 = rf(message)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(response.Response)
		}
	}

	return r0
}
