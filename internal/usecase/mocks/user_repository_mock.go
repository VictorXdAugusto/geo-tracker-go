package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
)

// MockUserRepository é um mock do UserRepository para testes
type MockUserRepository struct {
	mock.Mock
}

// Save mock
func (m *MockUserRepository) Save(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// FindByID mock
func (m *MockUserRepository) FindByID(ctx context.Context, id entity.UserID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// FindByEmail mock
func (m *MockUserRepository) FindByEmail(ctx context.Context, email entity.Email) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// Exists mock
func (m *MockUserRepository) Exists(ctx context.Context, id entity.UserID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// Delete mock
func (m *MockUserRepository) Delete(ctx context.Context, id entity.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindAll mock
func (m *MockUserRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}
