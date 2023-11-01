// Code generated by mockery v2.36.0. DO NOT EDIT.

package logger

import mock "github.com/stretchr/testify/mock"

// MockPolicy is an autogenerated mock type for the Policy type
type MockPolicy struct {
	mock.Mock
}

// Allow provides a mock function with given fields:
func (_m *MockPolicy) Allow() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NewMockPolicy creates a new instance of MockPolicy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPolicy(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPolicy {
	mock := &MockPolicy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
