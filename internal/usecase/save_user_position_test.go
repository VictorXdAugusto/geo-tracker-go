package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/internal/usecase/mocks"
)

// SaveUserPositionUseCaseTestSuite define a suite de testes para SaveUserPositionUseCase
type SaveUserPositionUseCaseTestSuite struct {
	suite.Suite
	userRepo       *mocks.MockUserRepository
	positionRepo   *mocks.MockPositionRepository
	eventPublisher *mocks.MockEventPublisher
	cache          *mocks.MockCache
	logger         *mocks.MockLogger
	useCase        *usecase.SaveUserPositionUseCase
	ctx            context.Context
	validUser      *entity.User
}

// SetupTest configura cada teste
func (suite *SaveUserPositionUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.positionRepo = new(mocks.MockPositionRepository)
	suite.eventPublisher = new(mocks.MockEventPublisher)
	suite.cache = new(mocks.MockCache)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewSaveUserPositionUseCase(
		suite.userRepo,
		suite.positionRepo,
		suite.eventPublisher,
		suite.cache,
		suite.logger,
	)
	suite.ctx = context.Background()

	// Criar usuário válido para reutilizar nos testes
	var err error
	suite.validUser, err = entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)
}

// TearDownTest limpa após cada teste
func (suite *SaveUserPositionUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.positionRepo.AssertExpectations(suite.T())
	suite.eventPublisher.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// addCacheInvalidationMocks adiciona mocks de invalidação de cache para testes de escrita
func (suite *SaveUserPositionUseCaseTestSuite) addCacheInvalidationMocks(userID string) {
	// Mocks para invalidação de cache (podem falhar sem quebrar o teste)
	suite.cache.On("Delete", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.Contains(key, userID)
	})).Return(nil).Maybe()

	// Mock para log de debug da invalidação do cache
	suite.logger.On("Debug", "Cache invalidation completed", mock.Anything).Return().Maybe()
}

// TestSaveUserPosition_Success testa salvamento bem-sucedido de posição
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_Success() {
	// Arrange
	now := time.Now()
	request := usecase.SaveUserPositionRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
		Timestamp: now,
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	// Adicionar mocks de invalidação de cache
	suite.addCacheInvalidationMocks(request.UserID)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(suite.validUser, nil)

	// Mock: buscar posição anterior (pode não existir)
	suite.positionRepo.On("FindCurrentByUserID", mock.Anything, *userID).
		Return(nil, errors.New("no previous position")).Maybe()

	// Mock: salvar posição com sucesso
	suite.positionRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Position")).
		Return(nil)

	// Mock: publicar evento com sucesso
	suite.eventPublisher.On("PublishPositionChanged", mock.Anything, mock.AnythingOfType("*events.Event")).
		Return(nil)

	// Mock: logs de sucesso
	suite.logger.On("Info", "Position saved successfully", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.NotEmpty(suite.T(), response.PositionID)
	assert.NotEmpty(suite.T(), response.SectorID)
	assert.Equal(suite.T(), "Position saved successfully", response.Message)
}

// TestSaveUserPosition_UserNotFound testa quando usuário não existe
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_UserNotFound() {
	// Arrange
	request := usecase.SaveUserPositionRequest{
		UserID:    "nonexistent",
		Latitude:  -23.550520,
		Longitude: -46.633309,
		Timestamp: time.Now(),
	}

	userID, err := entity.NewUserID("nonexistent")
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

// TestSaveUserPosition_InvalidCoordinates testa com coordenadas inválidas
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_InvalidCoordinates() {
	testCases := []struct {
		name      string
		latitude  float64
		longitude float64
		wantErr   string
	}{
		{
			name:      "latitude muito alta",
			latitude:  91.0,
			longitude: -46.633309,
			wantErr:   "invalid coordinates",
		},
		{
			name:      "latitude muito baixa",
			latitude:  -91.0,
			longitude: -46.633309,
			wantErr:   "invalid coordinates",
		},
		{
			name:      "longitude muito alta",
			latitude:  -23.550520,
			longitude: 181.0,
			wantErr:   "invalid coordinates",
		},
		{
			name:      "longitude muito baixa",
			latitude:  -23.550520,
			longitude: -181.0,
			wantErr:   "invalid coordinates",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Arrange
			request := usecase.SaveUserPositionRequest{
				UserID:    "user123",
				Latitude:  tc.latitude,
				Longitude: tc.longitude,
				Timestamp: time.Now(),
			}

			userID, err := entity.NewUserID("user123")
			suite.Require().NoError(err)

			// Mock: usuário existe (precisa passar validação de usuário primeiro)
			suite.userRepo.On("FindByID", mock.Anything, *userID).
				Return(suite.validUser, nil)

			// Mock: log de erro esperado
			suite.logger.On("Error", "Invalid coordinates", mock.Anything).
				Return()

			// Act
			response, err := suite.useCase.Execute(suite.ctx, request)

			// Assert
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), response)
			assert.Contains(suite.T(), err.Error(), tc.wantErr)
		})
	}
}

// TestSaveUserPosition_RepositoryError testa erro ao salvar no repositório
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_RepositoryError() {
	// Arrange
	request := usecase.SaveUserPositionRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
		Timestamp: time.Now(),
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	repositoryError := errors.New("database connection failed")

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(suite.validUser, nil)

	// Mock: buscar posição anterior (pode não existir)
	suite.positionRepo.On("FindCurrentByUserID", mock.Anything, *userID).
		Return(nil, errors.New("no previous position"))

	// Mock: erro ao salvar posição
	suite.positionRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Position")).
		Return(repositoryError)

	// Mock: log de erro
	suite.logger.On("Error", "Failed to save position", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "failed to save position")
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// TestSaveUserPosition_EventPublishError testa erro ao publicar evento
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_EventPublishError() {
	// Arrange
	request := usecase.SaveUserPositionRequest{
		UserID:    "user123",
		Latitude:  -23.550520,
		Longitude: -46.633309,
		Timestamp: time.Now(),
	}

	userID, err := entity.NewUserID("user123")
	suite.Require().NoError(err)

	eventError := errors.New("event publisher failed")

	// Adicionar mocks de invalidação de cache
	suite.addCacheInvalidationMocks(request.UserID)

	// Mock: usuário existe
	suite.userRepo.On("FindByID", mock.Anything, *userID).
		Return(suite.validUser, nil)

	// Mock: buscar posição anterior (pode não existir)
	suite.positionRepo.On("FindCurrentByUserID", mock.Anything, *userID).
		Return(nil, errors.New("no previous position"))

	// Mock: salvar posição com sucesso
	suite.positionRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Position")).
		Return(nil)

	// Mock: erro ao publicar evento
	suite.eventPublisher.On("PublishPositionChanged", mock.Anything, mock.AnythingOfType("*events.Event")).
		Return(eventError)

	// Mock: logs - sucesso ao salvar e erro no evento
	suite.logger.On("Info", "Position saved successfully", mock.Anything).
		Return()
	suite.logger.On("Error", "Failed to publish position changed event",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	// NOTE: Dependendo da implementação, erro no evento pode ou não falhar todo o processo
	// Assumindo que position é salva mesmo com erro no evento
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
}

// TestSaveUserPosition_InvalidUserID testa com ID de usuário inválido
func (suite *SaveUserPositionUseCaseTestSuite) TestSaveUserPosition_InvalidUserID() {
	// Arrange
	request := usecase.SaveUserPositionRequest{
		UserID:    "", // ID vazio
		Latitude:  -23.550520,
		Longitude: -46.633309,
		Timestamp: time.Now(),
	}

	// Mock: log de erro pode ser chamado
	suite.logger.On("Error", "Invalid user ID", mock.Anything).
		Return().Maybe()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "invalid user")
}

// TestNewSaveUserPositionUseCase testa o construtor
func (suite *SaveUserPositionUseCaseTestSuite) TestNewSaveUserPositionUseCase() {
	// Act
	uc := usecase.NewSaveUserPositionUseCase(
		suite.userRepo,
		suite.positionRepo,
		suite.eventPublisher,
		suite.cache,
		suite.logger,
	)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestSaveUserPositionUseCase executa toda a suite de testes
func TestSaveUserPositionUseCase(t *testing.T) {
	suite.Run(t, new(SaveUserPositionUseCaseTestSuite))
}
