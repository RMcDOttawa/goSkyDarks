// Code generated by MockGen. DO NOT EDIT.
// Source: goskydarks/delay (interfaces: DelayService)

// Package delay is a generated GoMock package.
package delay

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockDelayService is a mock of DelayService interface.
type MockDelayService struct {
	ctrl     *gomock.Controller
	recorder *MockDelayServiceMockRecorder
}

// MockDelayServiceMockRecorder is the mock recorder for MockDelayService.
type MockDelayServiceMockRecorder struct {
	mock *MockDelayService
}

// NewMockDelayService creates a new mock instance.
func NewMockDelayService(ctrl *gomock.Controller) *MockDelayService {
	mock := &MockDelayService{ctrl: ctrl}
	mock.recorder = &MockDelayServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDelayService) EXPECT() *MockDelayServiceMockRecorder {
	return m.recorder
}

// DelayDuration mocks base method.
func (m *MockDelayService) DelayDuration(arg0 int) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelayDuration", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DelayDuration indicates an expected call of DelayDuration.
func (mr *MockDelayServiceMockRecorder) DelayDuration(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelayDuration", reflect.TypeOf((*MockDelayService)(nil).DelayDuration), arg0)
}

// DelayUntil mocks base method.
func (m *MockDelayService) DelayUntil(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelayUntil", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DelayUntil indicates an expected call of DelayUntil.
func (mr *MockDelayServiceMockRecorder) DelayUntil(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelayUntil", reflect.TypeOf((*MockDelayService)(nil).DelayUntil), arg0)
}
