// Code generated by mockery v2.37.1. DO NOT EDIT.

package controllermock

import (
	context "context"

	controller "github.com/hashicorp/consul/internal/controller"
	mock "github.com/stretchr/testify/mock"

	pbresource "github.com/hashicorp/consul/proto-public/pbresource/v1"
)

// DependencyTransform is an autogenerated mock type for the DependencyTransform type
type DependencyTransform struct {
	mock.Mock
}

type DependencyTransform_Expecter struct {
	mock *mock.Mock
}

func (_m *DependencyTransform) EXPECT() *DependencyTransform_Expecter {
	return &DependencyTransform_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, rt, res
func (_m *DependencyTransform) Execute(ctx context.Context, rt controller.Runtime, res *pbresource.Resource) ([]*pbresource.Resource, error) {
	ret := _m.Called(ctx, rt, res)

	var r0 []*pbresource.Resource
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, controller.Runtime, *pbresource.Resource) ([]*pbresource.Resource, error)); ok {
		return rf(ctx, rt, res)
	}
	if rf, ok := ret.Get(0).(func(context.Context, controller.Runtime, *pbresource.Resource) []*pbresource.Resource); ok {
		r0 = rf(ctx, rt, res)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pbresource.Resource)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, controller.Runtime, *pbresource.Resource) error); ok {
		r1 = rf(ctx, rt, res)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DependencyTransform_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type DependencyTransform_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - rt controller.Runtime
//   - res *pbresource.Resource
func (_e *DependencyTransform_Expecter) Execute(ctx interface{}, rt interface{}, res interface{}) *DependencyTransform_Execute_Call {
	return &DependencyTransform_Execute_Call{Call: _e.mock.On("Execute", ctx, rt, res)}
}

func (_c *DependencyTransform_Execute_Call) Run(run func(ctx context.Context, rt controller.Runtime, res *pbresource.Resource)) *DependencyTransform_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(controller.Runtime), args[2].(*pbresource.Resource))
	})
	return _c
}

func (_c *DependencyTransform_Execute_Call) Return(_a0 []*pbresource.Resource, _a1 error) *DependencyTransform_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DependencyTransform_Execute_Call) RunAndReturn(run func(context.Context, controller.Runtime, *pbresource.Resource) ([]*pbresource.Resource, error)) *DependencyTransform_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewDependencyTransform creates a new instance of DependencyTransform. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDependencyTransform(t interface {
	mock.TestingT
	Cleanup(func())
}) *DependencyTransform {
	mock := &DependencyTransform{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
