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

// GetPositionHistoryUseCaseTestSuite define a suite de testes para GetPositionHistoryUseCase
type GetPositionHistoryUseCaseTestSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	positionRepo *mocks.MockPositionRepository
	cache        *mocks.MockCache
	logger       *mocks.MockLogger
	useCase      *usecase.GetPositionHistoryUseCase
	ctx          context.Context
}

// SetupTest configura cada teste
func (suite *GetPositionHistoryUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.positionRepo = new(mocks.MockPositionRepository)
	suite.cache = new(mocks.MockCache)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewGetPositionHistoryUseCase(suite.userRepo, suite.positionRepo, suite.cache, suite.logger)
	suite.ctx = context.Background()
}

// TearDownTest limpa após cada teste
func (suite *GetPositionHistoryUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.positionRepo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// addCacheMissMocks adiciona mocks padrão de cache miss para testes de leitura
func (suite *GetPositionHistoryUseCaseTestSuite) addCacheMissMocks(userID string, limit int) {
	suite.cache.On("GetCachedUserHistory", mock.Anything, userID, limit, mock.Anything).
		Return(errors.New("cache miss")).Maybe()
	suite.cache.On("CacheUserHistory", mock.Anything, userID, limit, mock.Anything).
		Return(nil).Maybe()
}

// TestGetPositionHistory_Success testa busca bem-sucedida
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_Success() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "user123",
		Limit:  10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Criar histórico de posições
	position1, err := entity.NewPosition("pos-1", *userID, -23.550520, -46.633309, time.Now().Add(-2*time.Hour))
	suite.Require().NoError(err)

	position2, err := entity.NewPosition("pos-2", *userID, -23.551000, -46.634000, time.Now().Add(-1*time.Hour))
	suite.Require().NoError(err)

	positions := []*entity.Position{position1, position2}

	// Mock: cache miss primeiro
	suite.cache.On("GetCachedUserHistory", mock.Anything, request.UserID, 10, mock.Anything).
		Return(errors.New("cache miss"))

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: histórico encontrado
	suite.positionRepo.On("FindHistoryByUserID", mock.Anything, *userID, 10).
		Return(positions, nil)

	// Mock: cachear o resultado
	suite.cache.On("CacheUserHistory", mock.Anything, request.UserID, 10, mock.Anything).
		Return(nil)

	// Mock: log de sucesso do banco de dados
	suite.logger.On("Info", "Position history retrieved from database", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.UserID)
	assert.Equal(suite.T(), "João Silva", response.UserName)
	assert.Equal(suite.T(), 2, response.Total)
	assert.Len(suite.T(), response.History, 2)
	assert.Equal(suite.T(), "pos-1", response.History[0].PositionID)
	assert.Equal(suite.T(), "pos-2", response.History[1].PositionID)
}

// TestGetPositionHistory_UserNotFound testa usuário não encontrado
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_UserNotFound() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "user123",
		Limit:  10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	// Mock: cache miss primeiro
	suite.cache.On("GetCachedUserHistory", mock.Anything, request.UserID, 10, mock.Anything).
		Return(errors.New("cache miss"))

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

// TestGetPositionHistory_RepositoryError testa erro do repositório
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_RepositoryError() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "user123",
		Limit:  10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	repoError := errors.New("database error")

	// Adicionar mocks de cache miss
	suite.addCacheMissMocks(request.UserID, request.Limit)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: erro no repositório
	suite.positionRepo.On("FindHistoryByUserID", mock.Anything, *userID, 10).
		Return(nil, repoError)

	// Mock: log de erro
	suite.logger.On("Error", "Failed to get position history", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestGetPositionHistory_EmptyHistory testa histórico vazio
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_EmptyHistory() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "user123",
		Limit:  10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Adicionar mocks de cache miss
	suite.addCacheMissMocks(request.UserID, request.Limit)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: histórico vazio
	suite.positionRepo.On("FindHistoryByUserID", mock.Anything, *userID, 10).
		Return([]*entity.Position{}, nil)

	// Mock: log de sucesso do banco de dados
	suite.logger.On("Info", "Position history retrieved from database", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.UserID)
	assert.Equal(suite.T(), 0, response.Total)
	assert.Empty(suite.T(), response.History)
}

// TestGetPositionHistory_InvalidUserID testa ID de usuário inválido
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_InvalidUserID() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "", // ID vazio é inválido
		Limit:  10,
	}

	// Adicionar mocks de cache miss (pode ser chamado mesmo com ID inválido)
	suite.addCacheMissMocks(request.UserID, request.Limit)

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

// TestGetPositionHistory_DefaultLimit testa limite padrão
func (suite *GetPositionHistoryUseCaseTestSuite) TestGetPositionHistory_DefaultLimit() {
	// Arrange
	request := usecase.GetPositionHistoryRequest{
		UserID: "user123",
		Limit:  0, // Deve usar limite padrão
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Adicionar mocks de cache miss (limite será convertido para 10)
	suite.addCacheMissMocks(request.UserID, 10)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: histórico com limite padrão (10)
	suite.positionRepo.On("FindHistoryByUserID", mock.Anything, *userID, 10).
		Return([]*entity.Position{}, nil)

	// Mock: log de sucesso do banco de dados
	suite.logger.On("Info", "Position history retrieved from database", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
}

// TestNewGetPositionHistoryUseCase testa o construtor
func (suite *GetPositionHistoryUseCaseTestSuite) TestNewGetPositionHistoryUseCase() {
	// Act
	uc := usecase.NewGetPositionHistoryUseCase(suite.userRepo, suite.positionRepo, suite.cache, suite.logger)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestGetPositionHistoryUseCase executa toda a suite de testes
func TestGetPositionHistoryUseCase(t *testing.T) {
	suite.Run(t, new(GetPositionHistoryUseCaseTestSuite))
}
