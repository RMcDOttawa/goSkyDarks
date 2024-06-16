// Code generated by MockGen. DO NOT EDIT.
// Source: goskydarks/theSkyX (interfaces: TheSkyDriver)

// Package theSkyX is a generated GoMock package.
package theSkyX

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTheSkyDriver is a mock of TheSkyDriver interface.
type MockTheSkyDriver struct {
	ctrl     *gomock.Controller
	recorder *MockTheSkyDriverMockRecorder
}

// MockTheSkyDriverMockRecorder is the mock recorder for MockTheSkyDriver.
type MockTheSkyDriverMockRecorder struct {
	mock *MockTheSkyDriver
}

// NewMockTheSkyDriver creates a new mock instance.
func NewMockTheSkyDriver(ctrl *gomock.Controller) *MockTheSkyDriver {
	mock := &MockTheSkyDriver{ctrl: ctrl}
	mock.recorder = &MockTheSkyDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTheSkyDriver) EXPECT() *MockTheSkyDriverMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockTheSkyDriver) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockTheSkyDriverMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTheSkyDriver)(nil).Close))
}

// Connect mocks base method.
func (m *MockTheSkyDriver) Connect(arg0 string, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockTheSkyDriverMockRecorder) Connect(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockTheSkyDriver)(nil).Connect), arg0, arg1)
}

// ConnectCamera mocks base method.
func (m *MockTheSkyDriver) ConnectCamera() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectCamera")
	ret0, _ := ret[0].(error)
	return ret0
}

// ConnectCamera indicates an expected call of ConnectCamera.
func (mr *MockTheSkyDriverMockRecorder) ConnectCamera() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectCamera", reflect.TypeOf((*MockTheSkyDriver)(nil).ConnectCamera))
}

// GetCameraTemperature mocks base method.
func (m *MockTheSkyDriver) GetCameraTemperature() (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCameraTemperature")
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCameraTemperature indicates an expected call of GetCameraTemperature.
func (mr *MockTheSkyDriverMockRecorder) GetCameraTemperature() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCameraTemperature", reflect.TypeOf((*MockTheSkyDriver)(nil).GetCameraTemperature))
}

// IsCaptureDone mocks base method.
func (m *MockTheSkyDriver) IsCaptureDone() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsCaptureDone")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsCaptureDone indicates an expected call of IsCaptureDone.
func (mr *MockTheSkyDriverMockRecorder) IsCaptureDone() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsCaptureDone", reflect.TypeOf((*MockTheSkyDriver)(nil).IsCaptureDone))
}

// MeasureDownloadTime mocks base method.
func (m *MockTheSkyDriver) MeasureDownloadTime(arg0 int) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MeasureDownloadTime", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MeasureDownloadTime indicates an expected call of MeasureDownloadTime.
func (mr *MockTheSkyDriverMockRecorder) MeasureDownloadTime(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MeasureDownloadTime", reflect.TypeOf((*MockTheSkyDriver)(nil).MeasureDownloadTime), arg0)
}

// StartBiasFrameCapture mocks base method.
func (m *MockTheSkyDriver) StartBiasFrameCapture(arg0 int, arg1 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartBiasFrameCapture", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartBiasFrameCapture indicates an expected call of StartBiasFrameCapture.
func (mr *MockTheSkyDriverMockRecorder) StartBiasFrameCapture(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartBiasFrameCapture", reflect.TypeOf((*MockTheSkyDriver)(nil).StartBiasFrameCapture), arg0, arg1)
}

// StartCooling mocks base method.
func (m *MockTheSkyDriver) StartCooling(arg0 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartCooling", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartCooling indicates an expected call of StartCooling.
func (mr *MockTheSkyDriverMockRecorder) StartCooling(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCooling", reflect.TypeOf((*MockTheSkyDriver)(nil).StartCooling), arg0)
}

// StartDarkFrameCapture mocks base method.
func (m *MockTheSkyDriver) StartDarkFrameCapture(arg0 int, arg1, arg2 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartDarkFrameCapture", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartDarkFrameCapture indicates an expected call of StartDarkFrameCapture.
func (mr *MockTheSkyDriverMockRecorder) StartDarkFrameCapture(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartDarkFrameCapture", reflect.TypeOf((*MockTheSkyDriver)(nil).StartDarkFrameCapture), arg0, arg1, arg2)
}

// StopCooling mocks base method.
func (m *MockTheSkyDriver) StopCooling() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopCooling")
	ret0, _ := ret[0].(error)
	return ret0
}

// StopCooling indicates an expected call of StopCooling.
func (mr *MockTheSkyDriverMockRecorder) StopCooling() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopCooling", reflect.TypeOf((*MockTheSkyDriver)(nil).StopCooling))
}
