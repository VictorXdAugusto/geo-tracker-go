package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vitao/geolocation-tracker/internal/domain/events"
)

// MockEventPublisher é um mock do events.Publisher para testes
type MockEventPublisher struct {
	mock.Mock
}

// Publish mock
func (m *MockEventPublisher) Publish(ctx context.Context, streamName string, event *events.Event) error {
	args := m.Called(ctx, streamName, event)
	return args.Error(0)
}

// PublishPositionChanged mock
func (m *MockEventPublisher) PublishPositionChanged(ctx context.Context, event *events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// PublishSectorChanged mock
func (m *MockEventPublisher) PublishSectorChanged(ctx context.Context, event *events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Close mock
func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}
