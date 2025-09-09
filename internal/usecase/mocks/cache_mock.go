package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/vitao/geolocation-tracker/internal/usecase"
)

// MockCache é um mock para o cache Redis que implementa CacheInterface
type MockCache struct {
	mock.Mock
}

// Verifica se implementa a interface
var _ usecase.CacheInterface = (*MockCache)(nil)

// Get implementa o método Get do cache
func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

// Set implementa o método Set do cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// Delete implementa o método Delete do cache
func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// CacheUserPosition implementa o método helper de cache de posição
func (m *MockCache) CacheUserPosition(ctx context.Context, userID string, position interface{}) error {
	args := m.Called(ctx, userID, position)
	return args.Error(0)
}

// GetCachedUserPosition implementa o método helper de busca de posição
func (m *MockCache) GetCachedUserPosition(ctx context.Context, userID string, dest interface{}) error {
	args := m.Called(ctx, userID, dest)
	return args.Error(0)
}

// CacheNearbyUsers implementa o método helper de cache de usuários próximos
func (m *MockCache) CacheNearbyUsers(ctx context.Context, lat, lng, radius float64, users interface{}) error {
	args := m.Called(ctx, lat, lng, radius, users)
	return args.Error(0)
}

// GetCachedNearbyUsers implementa o método helper de busca de usuários próximos
func (m *MockCache) GetCachedNearbyUsers(ctx context.Context, lat, lng, radius float64, dest interface{}) error {
	args := m.Called(ctx, lat, lng, radius, dest)
	return args.Error(0)
}

// CacheUserHistory implementa o método helper de cache de histórico
func (m *MockCache) CacheUserHistory(ctx context.Context, userID string, limit int, history interface{}) error {
	args := m.Called(ctx, userID, limit, history)
	return args.Error(0)
}

// GetCachedUserHistory implementa o método helper de busca de histórico
func (m *MockCache) GetCachedUserHistory(ctx context.Context, userID string, limit int, dest interface{}) error {
	args := m.Called(ctx, userID, limit, dest)
	return args.Error(0)
}

// InvalidateUserCaches implementa o método de invalidação de caches do usuário
func (m *MockCache) InvalidateUserCaches(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
