// Code generated by MockGen. DO NOT EDIT.
// Source: internal/model/callbacks/callback_handlers.go

// Package mock_callbacks is a generated GoMock package.
package mock_callbacks

import (
	gomock "github.com/golang/mock/gomock"
)

// MockCallbackHandler is a mock of CallbackHandler interface.
type MockCallbackHandler struct {
	ctrl     *gomock.Controller
	recorder *MockCallbackHandlerMockRecorder
}

// MockCallbackHandlerMockRecorder is the mock recorder for MockCallbackHandler.
type MockCallbackHandlerMockRecorder struct {
	mock *MockCallbackHandler
}

// NewMockCallbackHandler creates a new mock instance.
func NewMockCallbackHandler(ctrl *gomock.Controller) *MockCallbackHandler {
	mock := &MockCallbackHandler{ctrl: ctrl}
	mock.recorder = &MockCallbackHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCallbackHandler) EXPECT() *MockCallbackHandlerMockRecorder {
	return m.recorder
}
