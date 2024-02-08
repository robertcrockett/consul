// Code generated by mockery v2.20.0. DO NOT EDIT.

package resource

import (
	internalresource "github.com/hashicorp/consul/internal/resource"
	mock "github.com/stretchr/testify/mock"

	pbresource "github.com/hashicorp/consul/proto-public/pbresource/v1"
)

// MockRegistry is an autogenerated mock type for the Registry type
type MockRegistry struct {
	mock.Mock
}

// Register provides a mock function with given fields: reg
func (_m *MockRegistry) Register(reg internalresource.Registration) {
	_m.Called(reg)
}

// Resolve provides a mock function with given fields: typ
func (_m *MockRegistry) Resolve(typ *pbresource.Type) (internalresource.Registration, bool) {
	ret := _m.Called(typ)

	var r0 internalresource.Registration
	var r1 bool
	if rf, ok := ret.Get(0).(func(*pbresource.Type) (internalresource.Registration, bool)); ok {
		return rf(typ)
	}
	if rf, ok := ret.Get(0).(func(*pbresource.Type) internalresource.Registration); ok {
		r0 = rf(typ)
	} else {
		r0 = ret.Get(0).(internalresource.Registration)
	}

	if rf, ok := ret.Get(1).(func(*pbresource.Type) bool); ok {
		r1 = rf(typ)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockRegistry interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockRegistry creates a new instance of MockRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockRegistry(t mockConstructorTestingTNewMockRegistry) *MockRegistry {
	mock := &MockRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
