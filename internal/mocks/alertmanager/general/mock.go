// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/prometheus/alertmanager/api/v2/client/general (interfaces: ClientService)
//
// Generated by this command:
//
//	mockgen -destination internal/mocks/alertmanager/general/mock.go github.com/prometheus/alertmanager/api/v2/client/general ClientService
//

// Package mock_general is a generated GoMock package.
package mock_general

import (
	reflect "reflect"

	runtime "github.com/go-openapi/runtime"
	general "github.com/prometheus/alertmanager/api/v2/client/general"
	gomock "go.uber.org/mock/gomock"
)

// MockClientService is a mock of ClientService interface.
type MockClientService struct {
	ctrl     *gomock.Controller
	recorder *MockClientServiceMockRecorder
}

// MockClientServiceMockRecorder is the mock recorder for MockClientService.
type MockClientServiceMockRecorder struct {
	mock *MockClientService
}

// NewMockClientService creates a new mock instance.
func NewMockClientService(ctrl *gomock.Controller) *MockClientService {
	mock := &MockClientService{ctrl: ctrl}
	mock.recorder = &MockClientServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientService) EXPECT() *MockClientServiceMockRecorder {
	return m.recorder
}

// GetStatus mocks base method.
func (m *MockClientService) GetStatus(arg0 *general.GetStatusParams, arg1 ...general.ClientOption) (*general.GetStatusOK, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetStatus", varargs...)
	ret0, _ := ret[0].(*general.GetStatusOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatus indicates an expected call of GetStatus.
func (mr *MockClientServiceMockRecorder) GetStatus(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockClientService)(nil).GetStatus), varargs...)
}

// SetTransport mocks base method.
func (m *MockClientService) SetTransport(arg0 runtime.ClientTransport) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTransport", arg0)
}

// SetTransport indicates an expected call of SetTransport.
func (mr *MockClientServiceMockRecorder) SetTransport(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTransport", reflect.TypeOf((*MockClientService)(nil).SetTransport), arg0)
}
