// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-kit/kit/metrics (interfaces: Gauge)

// Package mock_metrics is a generated GoMock package.
package mock_metrics

import (
	reflect "reflect"

	metrics "github.com/go-kit/kit/metrics"
	gomock "github.com/golang/mock/gomock"
)

// MockGauge is a mock of Gauge interface.
type MockGauge struct {
	ctrl     *gomock.Controller
	recorder *MockGaugeMockRecorder
}

// MockGaugeMockRecorder is the mock recorder for MockGauge.
type MockGaugeMockRecorder struct {
	mock *MockGauge
}

// NewMockGauge creates a new mock instance.
func NewMockGauge(ctrl *gomock.Controller) *MockGauge {
	mock := &MockGauge{ctrl: ctrl}
	mock.recorder = &MockGaugeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGauge) EXPECT() *MockGaugeMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockGauge) Add(arg0 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", arg0)
}

// Add indicates an expected call of Add.
func (mr *MockGaugeMockRecorder) Add(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockGauge)(nil).Add), arg0)
}

// Set mocks base method.
func (m *MockGauge) Set(arg0 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Set", arg0)
}

// Set indicates an expected call of Set.
func (mr *MockGaugeMockRecorder) Set(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockGauge)(nil).Set), arg0)
}

// With mocks base method.
func (m *MockGauge) With(arg0 ...string) metrics.Gauge {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(metrics.Gauge)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockGaugeMockRecorder) With(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockGauge)(nil).With), arg0...)
}
