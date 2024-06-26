// Code generated by MockGen. DO NOT EDIT.
// Source: goskydarks/session (interfaces: StateFileService)

// Package session is a generated GoMock package.
package session

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStateFileService is a mock of StateFileService interface.
type MockStateFileService struct {
	ctrl     *gomock.Controller
	recorder *MockStateFileServiceMockRecorder
}

// MockStateFileServiceMockRecorder is the mock recorder for MockStateFileService.
type MockStateFileServiceMockRecorder struct {
	mock *MockStateFileService
}

// NewMockStateFileService creates a new mock instance.
func NewMockStateFileService(ctrl *gomock.Controller) *MockStateFileService {
	mock := &MockStateFileService{ctrl: ctrl}
	mock.recorder = &MockStateFileServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStateFileService) EXPECT() *MockStateFileServiceMockRecorder {
	return m.recorder
}

// ReadStateFile mocks base method.
func (m *MockStateFileService) ReadStateFile() (*CapturePlan, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadStateFile")
	ret0, _ := ret[0].(*CapturePlan)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadStateFile indicates an expected call of ReadStateFile.
func (mr *MockStateFileServiceMockRecorder) ReadStateFile() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadStateFile", reflect.TypeOf((*MockStateFileService)(nil).ReadStateFile))
}

// SavePlanToFile mocks base method.
func (m *MockStateFileService) SavePlanToFile(arg0 *CapturePlan) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SavePlanToFile", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SavePlanToFile indicates an expected call of SavePlanToFile.
func (mr *MockStateFileServiceMockRecorder) SavePlanToFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SavePlanToFile", reflect.TypeOf((*MockStateFileService)(nil).SavePlanToFile), arg0)
}

// UpdatePlanFromFile mocks base method.
func (m *MockStateFileService) UpdatePlanFromFile(arg0 *CapturePlan) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePlanFromFile", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePlanFromFile indicates an expected call of UpdatePlanFromFile.
func (mr *MockStateFileServiceMockRecorder) UpdatePlanFromFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePlanFromFile", reflect.TypeOf((*MockStateFileService)(nil).UpdatePlanFromFile), arg0)
}
