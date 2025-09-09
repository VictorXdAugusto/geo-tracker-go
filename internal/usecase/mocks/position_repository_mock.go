package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
)

// MockPositionRepository é um mock do PositionRepository para testes
type MockPositionRepository struct {
	mock.Mock
}

// Save mock
func (m *MockPositionRepository) Save(ctx context.Context, position *entity.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

// FindByID mock
func (m *MockPositionRepository) FindByID(ctx context.Context, id entity.PositionID) (*entity.Position, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Position), args.Error(1)
}

// FindCurrentByUserID mock
func (m *MockPositionRepository) FindCurrentByUserID(ctx context.Context, userID entity.UserID) (*entity.Position, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Position), args.Error(1)
}

// FindHistoryByUserID mock
func (m *MockPositionRepository) FindHistoryByUserID(ctx context.Context, userID entity.UserID, limit int) ([]*entity.Position, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Position), args.Error(1)
}

// FindNearby mock
func (m *MockPositionRepository) FindNearby(ctx context.Context, coord *valueobject.Coordinate, radiusMeters float64, limit int) ([]*entity.Position, error) {
	args := m.Called(ctx, coord, radiusMeters, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Position), args.Error(1)
}

// FindInSector mock
func (m *MockPositionRepository) FindInSector(ctx context.Context, sector *valueobject.Sector) ([]*entity.Position, error) {
	args := m.Called(ctx, sector)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Position), args.Error(1)
}

// FindInSectors mock
func (m *MockPositionRepository) FindInSectors(ctx context.Context, sectors []*valueobject.Sector) ([]*entity.Position, error) {
	args := m.Called(ctx, sectors)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Position), args.Error(1)
}

// UpdateCurrentPosition mock
func (m *MockPositionRepository) UpdateCurrentPosition(ctx context.Context, position *entity.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

// DeleteOldPositions mock
func (m *MockPositionRepository) DeleteOldPositions(ctx context.Context, olderThan *valueobject.Timestamp) (int, error) {
	args := m.Called(ctx, olderThan)
	return args.Int(0), args.Error(1)
}
