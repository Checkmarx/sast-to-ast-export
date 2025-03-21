// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/checkmarxDev/ast-sast-export/internal/app/interfaces (interfaces: MethodLineRepo)
//
// Generated by this command:
//
//	mockgen -package mock_app_method_line -destination test/mocks/app/method_line/mock_provider.go github.com/checkmarxDev/ast-sast-export/internal/app/interfaces MethodLineRepo
//

// Package mock_app_method_line is a generated GoMock package.
package mock_app_method_line

import (
	reflect "reflect"

	interfaces "github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	gomock "go.uber.org/mock/gomock"
)

// MockMethodLineRepo is a mock of MethodLineRepo interface.
type MockMethodLineRepo struct {
	ctrl     *gomock.Controller
	recorder *MockMethodLineRepoMockRecorder
}

// MockMethodLineRepoMockRecorder is the mock recorder for MockMethodLineRepo.
type MockMethodLineRepoMockRecorder struct {
	mock *MockMethodLineRepo
}

// NewMockMethodLineRepo creates a new mock instance.
func NewMockMethodLineRepo(ctrl *gomock.Controller) *MockMethodLineRepo {
	mock := &MockMethodLineRepo{ctrl: ctrl}
	mock.recorder = &MockMethodLineRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMethodLineRepo) EXPECT() *MockMethodLineRepoMockRecorder {
	return m.recorder
}

// GetMethodLines mocks base method.
func (m *MockMethodLineRepo) GetMethodLines(arg0, arg1, arg2 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMethodLines", arg0, arg1, arg2)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMethodLines indicates an expected call of GetMethodLines.
func (mr *MockMethodLineRepoMockRecorder) GetMethodLines(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMethodLines", reflect.TypeOf((*MockMethodLineRepo)(nil).GetMethodLines), arg0, arg1, arg2)
}

// GetMethodLinesByPath mocks base method.
func (m *MockMethodLineRepo) GetMethodLinesByPath(arg0, arg1 string) ([]*interfaces.ResultPath, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMethodLinesByPath", arg0, arg1)
	ret0, _ := ret[0].([]*interfaces.ResultPath)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMethodLinesByPath indicates an expected call of GetMethodLinesByPath.
func (mr *MockMethodLineRepoMockRecorder) GetMethodLinesByPath(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMethodLinesByPath", reflect.TypeOf((*MockMethodLineRepo)(nil).GetMethodLinesByPath), arg0, arg1)
}
