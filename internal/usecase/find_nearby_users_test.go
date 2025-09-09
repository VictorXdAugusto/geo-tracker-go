package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/internal/usecase/mocks"
)

// FindNearbyUsersUseCaseTestSuite define a suite de testes para FindNearbyUsersUseCase
type FindNearbyUsersUseCaseTestSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	positionRepo *mocks.MockPositionRepository
	logger       *mocks.MockLogger
	useCase      *usecase.FindNearbyUsersUseCase
	ctx          context.Context
}

// SetupTest configura cada teste
func (suite *FindNearbyUsersUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.positionRepo = new(mocks.MockPositionRepository)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewFindNearbyUsersUseCase(suite.userRepo, suite.positionRepo, suite.logger)
	suite.ctx = context.Background()
}

// TearDownTest limpa após cada teste
func (suite *FindNearbyUsersUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.positionRepo.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// TestFindNearbyUsers_Success testa busca bem-sucedida
func (suite *FindNearbyUsersUseCaseTestSuite) TestFindNearbyUsers_Success() {
	// Arrange
	request := usecase.FindNearbyUsersRequest{
		UserID:     "user123",
		Latitude:   -23.550520,
		Longitude:  -46.633309,
		RadiusM:    1000.0,
		MaxResults: 10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: encontrar posições próximas - O use case chama com maxResults+1 = 11
	positions := []*entity.Position{} // Lista vazia para simplificar
	suite.positionRepo.On("FindNearby", mock.Anything, mock.Anything, 1000.0, 11).
		Return(positions, nil)

	// Mock: log de sucesso
	suite.logger.On("Info", "Nearby users search completed", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), 0, response.TotalFound)
	assert.Empty(suite.T(), response.NearbyUsers)
}

// TestFindNearbyUsers_InvalidCoordinates testa com coordenadas inválidas
func (suite *FindNearbyUsersUseCaseTestSuite) TestFindNearbyUsers_InvalidCoordinates() {
	// Arrange
	request := usecase.FindNearbyUsersRequest{
		UserID:     "user123",
		Latitude:   91.0, // Inválida
		Longitude:  -46.633309,
		RadiusM:    1000.0,
		MaxResults: 10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: log de erro pode ser chamado
	suite.logger.On("Error", "Invalid search coordinates", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// TestFindNearbyUsers_RepositoryError testa erro do repositório
func (suite *FindNearbyUsersUseCaseTestSuite) TestFindNearbyUsers_RepositoryError() {
	// Arrange
	request := usecase.FindNearbyUsersRequest{
		UserID:     "user123",
		Latitude:   -23.550520,
		Longitude:  -46.633309,
		RadiusM:    1000.0,
		MaxResults: 10,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	repoError := errors.New("database error")

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: erro no repositório - O use case chama com maxResults+1 = 11
	suite.positionRepo.On("FindNearby", mock.Anything, mock.Anything, 1000.0, 11).
		Return(nil, repoError)

	// Mock: log de erro
	suite.logger.On("Error", "Failed to find nearby positions", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestNewFindNearbyUsersUseCase testa o construtor
func (suite *FindNearbyUsersUseCaseTestSuite) TestNewFindNearbyUsersUseCase() {
	// Act
	uc := usecase.NewFindNearbyUsersUseCase(suite.userRepo, suite.positionRepo, suite.logger)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestFindNearbyUsersUseCase executa toda a suite de testes
func TestFindNearbyUsersUseCase(t *testing.T) {
	suite.Run(t, new(FindNearbyUsersUseCaseTestSuite))
}
