// Code generated by mockery v1.0.0
package mocks

import entity "github.com/asiragusa/wschat/entity"
import mock "github.com/stretchr/testify/mock"

import time "time"

// TokenGenerator is an autogenerated mock type for the TokenGenerator type
type TokenGenerator struct {
	mock.Mock
}

// GenerateToken provides a mock function with given fields: _a0, _a1
func (_m *TokenGenerator) GenerateToken(_a0 entity.User, _a1 time.Duration) (string, error) {
	ret := _m.Called(_a0, _a1)

	var r0 string
	if rf, ok := ret.Get(0).(func(entity.User, time.Duration) string); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(entity.User, time.Duration) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateToken provides a mock function with given fields: _a0
func (_m *TokenGenerator) ValidateToken(_a0 string) (*entity.User, error) {
	ret := _m.Called(_a0)

	var r0 *entity.User
	if rf, ok := ret.Get(0).(func(string) *entity.User); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
