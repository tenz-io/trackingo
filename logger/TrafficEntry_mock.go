// Code generated by mockery v2.36.0. DO NOT EDIT.

package logger

import mock "github.com/stretchr/testify/mock"

// MockTrafficEntry is an autogenerated mock type for the TrafficEntry type
type MockTrafficEntry struct {
	mock.Mock
}

// Data provides a mock function with given fields: traffic
func (_m *MockTrafficEntry) Data(traffic *Traffic) {
	_m.Called(traffic)
}

// DataWith provides a mock function with given fields: traffic, fields
func (_m *MockTrafficEntry) DataWith(traffic *Traffic, fields Fields) {
	_m.Called(traffic, fields)
}

// WithFields provides a mock function with given fields: fields
func (_m *MockTrafficEntry) WithFields(fields Fields) TrafficEntry {
	ret := _m.Called(fields)

	var r0 TrafficEntry
	if rf, ok := ret.Get(0).(func(Fields) TrafficEntry); ok {
		r0 = rf(fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(TrafficEntry)
		}
	}

	return r0
}

// WithIgnores provides a mock function with given fields: ignores
func (_m *MockTrafficEntry) WithIgnores(ignores ...string) TrafficEntry {
	_va := make([]interface{}, len(ignores))
	for _i := range ignores {
		_va[_i] = ignores[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 TrafficEntry
	if rf, ok := ret.Get(0).(func(...string) TrafficEntry); ok {
		r0 = rf(ignores...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(TrafficEntry)
		}
	}

	return r0
}

// WithPolicy provides a mock function with given fields: policy
func (_m *MockTrafficEntry) WithPolicy(policy Policy) TrafficEntry {
	ret := _m.Called(policy)

	var r0 TrafficEntry
	if rf, ok := ret.Get(0).(func(Policy) TrafficEntry); ok {
		r0 = rf(policy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(TrafficEntry)
		}
	}

	return r0
}

// WithTracing provides a mock function with given fields: requestId
func (_m *MockTrafficEntry) WithTracing(requestId string) TrafficEntry {
	ret := _m.Called(requestId)

	var r0 TrafficEntry
	if rf, ok := ret.Get(0).(func(string) TrafficEntry); ok {
		r0 = rf(requestId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(TrafficEntry)
		}
	}

	return r0
}

// NewMockTrafficEntry creates a new instance of MockTrafficEntry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTrafficEntry(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTrafficEntry {
	mock := &MockTrafficEntry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
