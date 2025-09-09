package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockLogger é um mock do Logger para testes
type MockLogger struct {
	mock.Mock
}

// Info mock
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, msg)
	args = append(args, fields...)
	m.Called(args...)
}

// Error mock
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, msg)
	args = append(args, fields...)
	m.Called(args...)
}

// Fatal mock
func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, msg)
	args = append(args, fields...)
	m.Called(args...)
}

// Debug mock
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, msg)
	args = append(args, fields...)
	m.Called(args...)
}

// Sync mock
func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}
