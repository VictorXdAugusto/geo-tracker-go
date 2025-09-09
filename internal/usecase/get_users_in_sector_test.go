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

// GetUsersInSectorUseCaseTestSuite define a suite de testes para GetUsersInSectorUseCase
type GetUsersInSectorUseCaseTestSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	positionRepo *mocks.MockPositionRepository
	cache        *mocks.MockCache
	logger       *mocks.MockLogger
	useCase      *usecase.GetUsersInSectorUseCase
	ctx          context.Context
}

// SetupTest configura cada teste
func (suite *GetUsersInSectorUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.positionRepo = new(mocks.MockPositionRepository)
	suite.cache = new(mocks.MockCache)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewGetUsersInSectorUseCase(suite.userRepo, suite.positionRepo, suite.cache, suite.logger)
	suite.ctx = context.Background()
}

// TearDownTest limpa após cada teste
func (suite *GetUsersInSectorUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.positionRepo.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// TestGetUsersInSector_Success testa busca bem-sucedida
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_Success() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Criar usuários no setor
	otherUserID, err := entity.NewUserID("user456")
	suite.Require().NoError(err)

	otherUser, err := entity.NewUser("user456", "Maria Santos", "maria@example.com")
	suite.Require().NoError(err)

	// Criar posições no mesmo setor (incluindo o usuário solicitante)
	selfPosition, err := entity.NewPosition("pos-self", *userID, -23.550520, -46.633309, time.Now().Add(-30*time.Minute))
	suite.Require().NoError(err)

	position1, err := entity.NewPosition("pos-1", *otherUserID, -23.550520, -46.633309, time.Now().Add(-30*time.Minute))
	suite.Require().NoError(err)

	positions := []*entity.Position{selfPosition, position1}

	// Mock: usuário solicitante existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: outros usuários do setor
	suite.userRepo.On("FindByID", mock.Anything, *otherUserID).
		Return(otherUser, nil)

	// Mock: posições no setor encontradas
	suite.positionRepo.On("FindInSector", mock.Anything, mock.Anything).
		Return(positions, nil)

	// Mock: log de sucesso
	suite.logger.On("Info", "Sector users search completed", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.RequestedBy.UserID)
	assert.Equal(suite.T(), "João Silva", response.RequestedBy.UserName)
	assert.Equal(suite.T(), 1, response.TotalFound)
	assert.Len(suite.T(), response.UsersInSector, 1)
	assert.Equal(suite.T(), "user456", response.UsersInSector[0].UserID)
	assert.Equal(suite.T(), "Maria Santos", response.UsersInSector[0].UserName)
}

// TestGetUsersInSector_UserNotFound testa usuário solicitante não encontrado
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_UserNotFound() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
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

// TestGetUsersInSector_RepositoryError testa erro do repositório
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_RepositoryError() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	repoError := errors.New("database error")

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: erro no repositório
	suite.positionRepo.On("FindInSector", mock.Anything, mock.Anything).
		Return(nil, repoError)

	// Mock: log de erro
	suite.logger.On("Error", "Failed to find positions in sector", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestGetUsersInSector_EmptySector testa setor vazio
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_EmptySector() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: setor vazio
	suite.positionRepo.On("FindInSector", mock.Anything, mock.Anything).
		Return([]*entity.Position{}, nil)

	// Mock: log de sucesso
	suite.logger.On("Info", "Sector users search completed", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	// O RequestedBy fica vazio quando o usuário não tem posição no setor
	assert.Equal(suite.T(), "", response.RequestedBy.UserID)
	assert.Equal(suite.T(), 0, response.TotalFound)
	assert.Empty(suite.T(), response.UsersInSector)
}

// TestGetUsersInSector_InvalidCoordinates testa coordenadas inválidas
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_InvalidCoordinates() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  91.0, // Inválida
		Longitude: -46.633309,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Mock: usuário existe (validação acontece antes das coordenadas)
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: log de erro para coordenadas inválidas
	suite.logger.On("Error", "Invalid coordinates", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// TestGetUsersInSector_InvalidUserID testa ID de usuário inválido
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_InvalidUserID() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "", // ID vazio é inválido
		Latitude:  -23.550520,
		Longitude: -46.633309,
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

// TestGetUsersInSector_ExcludeSelf testa que o usuário solicitante é excluído dos resultados
func (suite *GetUsersInSectorUseCaseTestSuite) TestGetUsersInSector_ExcludeSelf() {
	// Arrange
	request := usecase.GetUsersInSectorRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	validUser, err := entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	// Criar posição do próprio usuário (deve ser excluída)
	selfPosition, err := entity.NewPosition("pos-123", *userID, -23.550520, -46.633309, time.Now().Add(-30*time.Minute))
	suite.Require().NoError(err)

	positions := []*entity.Position{selfPosition}

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(validUser, nil)

	// Mock: posições incluem a do próprio usuário (que deve ser filtrada)
	suite.positionRepo.On("FindInSector", mock.Anything, mock.Anything).
		Return(positions, nil)

	// Mock: log de sucesso
	suite.logger.On("Info", "Sector users search completed", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), 0, response.TotalFound) // Próprio usuário é excluído
	assert.Empty(suite.T(), response.UsersInSector)
}

// TestNewGetUsersInSectorUseCase testa o construtor
func (suite *GetUsersInSectorUseCaseTestSuite) TestNewGetUsersInSectorUseCase() {
	// Act
	uc := usecase.NewGetUsersInSectorUseCase(suite.userRepo, suite.positionRepo, suite.cache, suite.logger)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestGetUsersInSectorUseCase executa toda a suite de testes
func TestGetUsersInSectorUseCase(t *testing.T) {
	suite.Run(t, new(GetUsersInSectorUseCaseTestSuite))
}
