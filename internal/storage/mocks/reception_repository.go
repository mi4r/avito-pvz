// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	storage "github.com/mi4r/avito-pvz/internal/storage"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// ReceptionRepository is an autogenerated mock type for the ReceptionRepository type
type ReceptionRepository struct {
	mock.Mock
}

// CloseReception provides a mock function with given fields: ctx, receptionID
func (_m *ReceptionRepository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	ret := _m.Called(ctx, receptionID)

	if len(ret) == 0 {
		panic("no return value specified for CloseReception")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, receptionID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateReception provides a mock function with given fields: ctx, pvzID
func (_m *ReceptionRepository) CreateReception(ctx context.Context, pvzID uuid.UUID) (storage.Reception, error) {
	ret := _m.Called(ctx, pvzID)

	if len(ret) == 0 {
		panic("no return value specified for CreateReception")
	}

	var r0 storage.Reception
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (storage.Reception, error)); ok {
		return rf(ctx, pvzID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) storage.Reception); ok {
		r0 = rf(ctx, pvzID)
	} else {
		r0 = ret.Get(0).(storage.Reception)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, pvzID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOpenReception provides a mock function with given fields: ctx, pvzID
func (_m *ReceptionRepository) GetOpenReception(ctx context.Context, pvzID uuid.UUID) (storage.Reception, error) {
	ret := _m.Called(ctx, pvzID)

	if len(ret) == 0 {
		panic("no return value specified for GetOpenReception")
	}

	var r0 storage.Reception
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (storage.Reception, error)); ok {
		return rf(ctx, pvzID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) storage.Reception); ok {
		r0 = rf(ctx, pvzID)
	} else {
		r0 = ret.Get(0).(storage.Reception)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, pvzID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewReceptionRepository creates a new instance of ReceptionRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReceptionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReceptionRepository {
	mock := &ReceptionRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
