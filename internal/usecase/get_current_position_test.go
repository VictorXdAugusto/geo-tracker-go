package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/internal/usecase/mocks"
)

// GetCurrentPositionUseCaseTestSuite define a suite de testes para GetCurrentPositionUseCase
type GetCurrentPositionUseCaseTestSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	positionRepo *mocks.MockPositionRepository
	logger       *mocks.MockLogger
	useCase      *usecase.GetCurrentPositionUseCase
	ctx          context.Context
}

// SetupTest configura cada teste
func (suite *GetCurrentPositionUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.positionRepo = new(mocks.MockPositionRepository)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewGetCurrentPositionUseCase(suite.userRepo, suite.positionRepo, suite.logger)
	suite.ctx = context.Background()
}

// TearDownTest limpa após cada teste
func (suite *GetCurrentPositionUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.positionRepo.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// TestGetCurrentPosition_Success testa busca bem-sucedida
func (suite *GetCurrentPositionUseCaseTestSuite) TestGetCurrentPosition_Success() {
	// Arrange
	request := usecase.GetCurrentPositionRequest{
		UserID: "user123",
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Criar position usando o construtor correto
	position, err := entity.NewPosition("pos-123", *userID, -23.550520, -46.633309, time.Now().Add(-1*time.Hour))
	suite.Require().NoError(err)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: posição atual encontrada
	suite.positionRepo.On("FindCurrentByUserID", mock.Anything, *userID).
		Return(position, nil)

	// Mock: log de sucesso
	suite.logger.On("Info", "Current position retrieved", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.UserID)
	assert.Equal(suite.T(), "João Silva", response.UserName)
	assert.Equal(suite.T(), "pos-123", response.PositionID)
	assert.Equal(suite.T(), -23.550520, response.Latitude)
	assert.Equal(suite.T(), -46.633309, response.Longitude)
	assert.NotEmpty(suite.T(), response.SectorID) // O setor é calculado automaticamente
}

// TestGetCurrentPosition_UserNotFound testa usuário não encontrado
func (suite *GetCurrentPositionUseCaseTestSuite) TestGetCurrentPosition_UserNotFound() {
	// Arrange
	request := usecase.GetCurrentPositionRequest{
		UserID: "user123",
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	// Mock: usuário não existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(nil, errors.New("user not found"))

	// Mock: log de erro
	suite.logger.On("Error", "User not found", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

// TestGetCurrentPosition_PositionNotFound testa posição não encontrada
func (suite *GetCurrentPositionUseCaseTestSuite) TestGetCurrentPosition_PositionNotFound() {
	// Arrange
	request := usecase.GetCurrentPositionRequest{
		UserID: "user123",
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: posição não encontrada
	suite.positionRepo.On("FindCurrentByUserID", mock.Anything, *userID).
		Return(nil, errors.New("position not found"))

	// Mock: log de erro
	suite.logger.On("Error", "Current position not found", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "position not found")
}

// TestGetCurrentPosition_InvalidUserID testa ID de usuário inválido
func (suite *GetCurrentPositionUseCaseTestSuite) TestGetCurrentPosition_InvalidUserID() {
	// Arrange
	request := usecase.GetCurrentPositionRequest{
		UserID: "", // ID vazio é inválido
	}

	// Mock: log de erro para ID inválido
	suite.logger.On("Error", "Invalid user ID", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// TestNewGetCurrentPositionUseCase testa o construtor
func (suite *GetCurrentPositionUseCaseTestSuite) TestNewGetCurrentPositionUseCase() {
	// Act
	uc := usecase.NewGetCurrentPositionUseCase(suite.userRepo, suite.positionRepo, suite.logger)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestGetCurrentPositionUseCase executa toda a suite de testes
func TestGetCurrentPositionUseCase(t *testing.T) {
	suite.Run(t, new(GetCurrentPositionUseCaseTestSuite))
}
